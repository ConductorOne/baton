package dotc1z

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func parseGrantPageToken(pageToken string, tokenType string) (int64, error) {
	id, err := strconv.ParseInt(pageToken, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s grants page token %q: %w", tokenType, pageToken, err)
	}
	return id, nil
}

// listExpandableGrantsInternal returns a page of grants with expansion
// metadata populated from the SQL columns directly, without unmarshalling
// the grant payload. Mode must be grantListModeExpansionNeedsOnly (the
// only caller is GrantStore.PendingExpansionPage). Consider removing the
// mode dispatch if no other caller ever appears.
func (c *C1File) listExpandableGrantsInternal(
	ctx context.Context,
	opts grantListOptions,
) ([]*expandableGrantDef, string, error) {
	if err := c.validateDb(ctx); err != nil {
		return nil, "", err
	}

	syncID, err := c.resolveSyncIDForInternalQuery(ctx, opts.SyncID)
	if err != nil {
		return nil, "", err
	}

	q := c.db.From(grants.Name()).Prepared(true)
	q = q.Select(
		"id",
		"external_id",
		"entitlement_id",
		"principal_resource_type_id",
		"principal_resource_id",
		"expansion",
		"needs_expansion",
	)
	q = q.Where(goqu.C("sync_id").Eq(syncID))
	q = q.Where(goqu.C("expansion").IsNotNull())
	if opts.Mode == grantListModeExpansionNeedsOnly {
		q = q.Where(goqu.C("needs_expansion").Eq(1))
	}

	if opts.PageToken != "" {
		id, err := parseGrantPageToken(opts.PageToken, "expandable")
		if err != nil {
			return nil, "", err
		}
		q = q.Where(goqu.C("id").Gte(id))
	}

	pageSize := opts.PageSize
	if pageSize > maxPageSize || pageSize == 0 {
		pageSize = maxPageSize
	}
	q = q.Order(goqu.C("id").Asc()).Limit(uint(pageSize + 1))

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, "", err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	defs := make([]*expandableGrantDef, 0, pageSize)
	var (
		count   uint32
		lastRow int64
	)
	for rows.Next() {
		count++
		if count > pageSize {
			break
		}

		var (
			rowID             int64
			externalID        string
			targetEntID       string
			principalRTID     string
			principalRID      string
			expansionBlob     []byte
			needsExpansionInt int
		)

		if err := rows.Scan(
			&rowID,
			&externalID,
			&targetEntID,
			&principalRTID,
			&principalRID,
			&expansionBlob,
			&needsExpansionInt,
		); err != nil {
			return nil, "", err
		}
		lastRow = rowID

		ge := &v2.GrantExpandable{}
		if err := proto.Unmarshal(expansionBlob, ge); err != nil {
			return nil, "", fmt.Errorf("invalid expansion data for %q: %w", externalID, err)
		}

		defs = append(defs, &expandableGrantDef{
			RowID:                   rowID,
			GrantExternalID:         externalID,
			TargetEntitlementID:     targetEntID,
			PrincipalResourceTypeID: principalRTID,
			PrincipalResourceID:     principalRID,
			SourceEntitlementIDs:    ge.GetEntitlementIds(),
			Shallow:                 ge.GetShallow(),
			ResourceTypeIDs:         ge.GetResourceTypeIds(),
			NeedsExpansion:          needsExpansionInt != 0,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	nextPageToken := ""
	if count > pageSize {
		nextPageToken = strconv.FormatInt(lastRow+1, 10)
	}
	return defs, nextPageToken, nil
}

func (c *C1File) listGrantsWithExpansionInternal(ctx context.Context, opts grantListOptions) (*internalGrantListResponse, error) {
	if err := c.validateDb(ctx); err != nil {
		return nil, err
	}

	syncID, err := c.resolveSyncIDForPayloadQuery(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting sync id for list grants with expansion: %w", err)
	}

	q := c.db.From(grants.Name()).Prepared(true).
		Select(
			"id",
			"data",
			"external_id",
			"entitlement_id",
			"resource_type_id",
			"resource_id",
			"principal_resource_type_id",
			"principal_resource_id",
			"expansion",
			"needs_expansion",
		)

	if resource := opts.Resource; resource != nil {
		q = q.Where(goqu.C("resource_id").Eq(resource.GetId().GetResource()))
		q = q.Where(goqu.C("resource_type_id").Eq(resource.GetId().GetResourceType()))
	}

	if syncID != "" {
		q = q.Where(goqu.C("sync_id").Eq(syncID))
	}

	if opts.ExpandableOnly {
		q = q.Where(goqu.C("expansion").IsNotNull())
	}
	if opts.PageToken != "" {
		id, err := parseGrantPageToken(opts.PageToken, "payload")
		if err != nil {
			return nil, err
		}
		q = q.Where(goqu.C("id").Gte(id))
	}

	pageSize := opts.PageSize
	if pageSize > maxPageSize || pageSize == 0 {
		pageSize = maxPageSize
	}
	q = q.Order(goqu.C("id").Asc()).Limit(uint(pageSize + 1))

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	unmarshalerOptions := proto.UnmarshalOptions{
		Merge:          true,
		DiscardUnknown: true,
	}

	var (
		rowID          int64
		grantData      sql.RawBytes
		externalIDRaw  sql.RawBytes
		entIDRaw       sql.RawBytes
		entRTRaw       sql.RawBytes
		entRRaw        sql.RawBytes
		principalRTRaw sql.RawBytes
		principalRRaw  sql.RawBytes
		expansionBlob  sql.RawBytes
		needsExp       int
		count          uint32
		lastRow        int64
	)
	var result []*internalGrantRow
	var pageGrants []*v2.Grant
	var pageKeys []grantJoinKeys

	for rows.Next() {
		count++
		if count > pageSize {
			break
		}

		if err := rows.Scan(
			&rowID,
			&grantData,
			&externalIDRaw,
			&entIDRaw,
			&entRTRaw,
			&entRRaw,
			&principalRTRaw,
			&principalRRaw,
			&expansionBlob,
			&needsExp,
		); err != nil {
			return nil, err
		}
		lastRow = rowID

		grant := &v2.Grant{}
		if err := unmarshalerOptions.Unmarshal(grantData, grant); err != nil {
			return nil, err
		}

		var expansion *expandableGrantDef
		if len(expansionBlob) > 0 {
			ge := &v2.GrantExpandable{}
			if err := proto.Unmarshal(expansionBlob, ge); err != nil {
				return nil, fmt.Errorf("failed to unmarshal grant expansion for %s: %w", string(externalIDRaw), err)
			}
			expansion = &expandableGrantDef{
				GrantExternalID:         string(externalIDRaw),
				TargetEntitlementID:     string(entIDRaw),
				PrincipalResourceTypeID: string(principalRTRaw),
				PrincipalResourceID:     string(principalRRaw),
				SourceEntitlementIDs:    ge.GetEntitlementIds(),
				Shallow:                 ge.GetShallow(),
				ResourceTypeIDs:         ge.GetResourceTypeIds(),
				NeedsExpansion:          needsExp != 0,
			}
		}

		result = append(result, &internalGrantRow{
			Grant:     grant,
			Expansion: expansion,
		})
		if grant.GetEntitlement() == nil || grant.GetPrincipal() == nil {
			pageGrants = append(pageGrants, grant)
			pageKeys = append(pageKeys, grantJoinKeys{
				EntitlementID:             string(entIDRaw),
				EntitlementResourceTypeID: string(entRTRaw),
				EntitlementResourceID:     string(entRRaw),
				PrincipalResourceTypeID:   string(principalRTRaw),
				PrincipalResourceID:       string(principalRRaw),
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(pageGrants) > 0 {
		hydrateGrants(pageGrants, pageKeys)
	}

	nextPageToken := ""
	if count > pageSize {
		nextPageToken = strconv.FormatInt(lastRow+1, 10)
	}

	return &internalGrantListResponse{
		Rows:          result,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) resolveSyncIDForPayloadQuery(ctx context.Context) (string, error) {
	if c.currentSyncID != "" {
		return c.currentSyncID, nil
	}
	if c.viewSyncID != "" {
		return c.viewSyncID, nil
	}
	latestSyncRun, err := c.getCachedViewSyncRun(ctx)
	if err != nil {
		return "", err
	}
	if latestSyncRun != nil {
		return latestSyncRun.ID, nil
	}
	return "", nil
}

func (c *C1File) resolveSyncIDForInternalQuery(ctx context.Context, forced string) (string, error) {
	switch {
	case forced != "":
		return forced, nil
	case c.currentSyncID != "":
		return c.currentSyncID, nil
	case c.viewSyncID != "":
		return c.viewSyncID, nil
	default:
		latest, err := c.getCachedViewSyncRun(ctx)
		if err != nil {
			return "", err
		}
		if latest == nil {
			return "", sql.ErrNoRows
		}
		return latest.ID, nil
	}
}
