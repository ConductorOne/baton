package dotc1z

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/uotel"
)

const grantsTableVersion = "1"
const grantsTableName = "grants"
const grantsTableSchema = `
create table if not exists %s (
    id integer primary key,
	resource_type_id text not null,
    resource_id text not null,
    entitlement_id text not null,
    principal_resource_type_id text not null,
    principal_resource_id text not null,
    external_id text not null,
    expansion blob,                             -- Serialized GrantExpandable proto; NULL if grant is not expandable.
    needs_expansion integer not null default 0, -- 1 if grant should be processed during expansion.
    data blob not null,
    sync_id text not null,
    discovered_at datetime not null
);
create index if not exists %s on %s (resource_type_id, resource_id);
create index if not exists %s on %s (principal_resource_type_id, principal_resource_id);
create index if not exists %s on %s (entitlement_id, principal_resource_type_id, principal_resource_id);
create unique index if not exists %s on %s (external_id, sync_id);`

var grants = (*grantsTable)(nil)

var _ tableDescriptor = (*grantsTable)(nil)

type grantsTable struct{}

func (r *grantsTable) Version() string {
	return grantsTableVersion
}

func (r *grantsTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), grantsTableName)
}

func (r *grantsTable) Schema() (string, []any) {
	return grantsTableSchema, []any{
		r.Name(),
		fmt.Sprintf("idx_grants_resource_type_id_resource_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_principal_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_entitlement_id_principal_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_external_sync_v%s", r.Version()),
		r.Name(),
	}
}

// isAlreadyExistsError returns true if err is a SQLite "duplicate column name" error.
func isAlreadyExistsError(err error) bool {
	return err != nil && strings.Contains(err.Error(), "duplicate column name")
}

func (r *grantsTable) Migrations(ctx context.Context, db *goqu.Database) error {
	// Add expansion column if missing (for older files).
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		"alter table %s add column expansion blob", r.Name(),
	)); err != nil && !isAlreadyExistsError(err) {
		return err
	}

	// Add needs_expansion column if missing.
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		"alter table %s add column needs_expansion integer not null default 0", r.Name(),
	)); err != nil && !isAlreadyExistsError(err) {
		return err
	}

	// Create partial index for efficient queries on expandable grants.
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		"create index if not exists %s on %s (sync_id) where expansion is not null",
		fmt.Sprintf("idx_grants_sync_expansion_v%s", r.Version()),
		r.Name(),
	)); err != nil {
		return err
	}

	// Create partial index for grants needing expansion processing.
	// Using a partial index (WHERE needs_expansion = 1) avoids polluting the query planner
	// for general grant queries — without this, SQLite may prefer this index over more
	// selective compound indexes like (entitlement_id, principal_resource_type_id, principal_resource_id).
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		"create index if not exists %s on %s (sync_id) where needs_expansion = 1",
		fmt.Sprintf("idx_grants_sync_needs_expansion_v%s", r.Version()),
		r.Name(),
	)); err != nil {
		return err
	}

	// Backfill expansion column from stored grant bytes.
	return backfillGrantExpansionColumn(ctx, db, r.Name())
}

func (c *C1File) ListGrants(ctx context.Context, request *v2.GrantsServiceListGrantsRequest) (*v2.GrantsServiceListGrantsResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrants")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	ret, nextPageToken, err := listGrantsGeneric(ctx, c, request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants: %w", err)
	}

	return v2.GrantsServiceListGrantsResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

// listGrantsGeneric pulls the grant identity columns inline so slim-blob
// rows can be hydrated without a second query.
func listGrantsGeneric(ctx context.Context, c *C1File, req listRequest) ([]*v2.Grant, string, error) {
	if err := c.validateDb(ctx); err != nil {
		return nil, "", err
	}

	reqSyncID, err := resolveSyncID(ctx, c, req)
	if err != nil {
		return nil, "", err
	}

	tableName := grants.Name()
	q := c.db.From(tableName).Prepared(true).Select(
		"id",
		"data",
		"entitlement_id",
		"resource_type_id",
		"resource_id",
		"principal_resource_type_id",
		"principal_resource_id",
	)

	// Filter predicates — mirrors listConnectorObjects.
	if resourceTypeReq, ok := req.(hasResourceTypeListRequest); ok {
		rt := resourceTypeReq.GetResourceTypeId()
		if rt != "" {
			q = q.Where(goqu.C("resource_type_id").Eq(rt))
		}
	}
	if resourceIdReq, ok := req.(hasResourceIdListRequest); ok {
		r := resourceIdReq.GetResourceId()
		if r != nil && r.GetResource() != "" {
			q = q.Where(goqu.C("resource_id").Eq(r.GetResource()))
			q = q.Where(goqu.C("resource_type_id").Eq(r.GetResourceType()))
		}
	}
	if resourceReq, ok := req.(hasResourceListRequest); ok {
		r := resourceReq.GetResource()
		if r != nil {
			q = q.Where(goqu.C("resource_id").Eq(r.GetId().GetResource()))
			q = q.Where(goqu.C("resource_type_id").Eq(r.GetId().GetResourceType()))
		}
	}
	if entitlementReq, ok := req.(hasEntitlementListRequest); ok {
		e := entitlementReq.GetEntitlement()
		if e != nil {
			q = q.Where(goqu.C("entitlement_id").Eq(e.GetId()))
		}
	}
	if principalIdReq, ok := req.(hasPrincipalIdListRequest); ok {
		p := principalIdReq.GetPrincipalId()
		if p != nil {
			q = q.Where(goqu.C("principal_resource_id").Eq(p.GetResource()))
			q = q.Where(goqu.C("principal_resource_type_id").Eq(p.GetResourceType()))
		}
	}
	if principalResourceTypeIDsReq, ok := req.(hasPrincipalResourceTypeIDsListRequest); ok {
		p := principalResourceTypeIDsReq.GetPrincipalResourceTypeIds()
		if len(p) > 0 {
			q = q.Where(goqu.C("principal_resource_type_id").In(p))
		}
	}

	if reqSyncID != "" {
		q = q.Where(goqu.C("sync_id").Eq(reqSyncID))
	}
	if req.GetPageToken() != "" {
		q = q.Where(goqu.C("id").Gte(req.GetPageToken()))
	}

	pageSize := req.GetPageSize()
	if pageSize > maxPageSize || pageSize == 0 {
		pageSize = maxPageSize
	}
	q = q.Order(goqu.C("id").Asc()).Limit(uint(pageSize + 1))

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, "", err
	}

	queryStart := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()
	if dur := time.Since(queryStart); dur > c.slowQueryThreshold {
		c.throttledWarnSlowQuery(ctx, query, dur)
	}

	unmarshal := proto.UnmarshalOptions{Merge: true, DiscardUnknown: true}

	var out []*v2.Grant
	var slimGrants []*v2.Grant
	var slimKeys []grantJoinKeys
	var (
		rowID          int
		data           sql.RawBytes
		entIDRaw       sql.RawBytes
		entRTRaw       sql.RawBytes
		entRRaw        sql.RawBytes
		principalRTRaw sql.RawBytes
		principalRRaw  sql.RawBytes
		count          uint32
		lastRow        int
	)
	for rows.Next() {
		count++
		if count > pageSize {
			break
		}
		if err := rows.Scan(&rowID, &data, &entIDRaw, &entRTRaw, &entRRaw, &principalRTRaw, &principalRRaw); err != nil {
			return nil, "", err
		}
		lastRow = rowID

		g := &v2.Grant{}
		if err := unmarshal.Unmarshal(data, g); err != nil {
			return nil, "", err
		}
		out = append(out, g)
		if g.GetEntitlement() == nil || g.GetPrincipal() == nil {
			slimGrants = append(slimGrants, g)
			slimKeys = append(slimKeys, grantJoinKeys{
				EntitlementID:             string(entIDRaw),
				EntitlementResourceTypeID: string(entRTRaw),
				EntitlementResourceID:     string(entRRaw),
				PrincipalResourceTypeID:   string(principalRTRaw),
				PrincipalResourceID:       string(principalRRaw),
			})
		}
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	if len(slimGrants) > 0 {
		hydrateGrants(slimGrants, slimKeys)
	}

	nextPageToken := ""
	if count > pageSize {
		nextPageToken = strconv.Itoa(lastRow + 1)
	}
	return out, nextPageToken, nil
}

func (c *C1File) GetGrant(ctx context.Context, request *reader_v2.GrantsReaderServiceGetGrantRequest) (*reader_v2.GrantsReaderServiceGetGrantResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.GetGrant")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	ret := &v2.Grant{}
	syncId, err := annotations.GetSyncIdFromAnnotations(request.GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error getting sync id from annotations for grant '%s': %w", request.GetGrantId(), err)
	}
	err = c.getConnectorObject(ctx, grants.Name(), request.GetGrantId(), syncId, ret)
	if err != nil {
		return nil, fmt.Errorf("error fetching grant '%s': %w", request.GetGrantId(), err)
	}

	// Re-resolve sync_id with the same cascade as getConnectorObject so
	// hydration hits the same row.
	if ret.GetEntitlement() == nil || ret.GetPrincipal() == nil {
		resolvedSyncID, rerr := c.resolveSyncIDForGrantGet(ctx, syncId)
		if rerr != nil {
			return nil, fmt.Errorf("error resolving sync id for grant '%s': %w", request.GetGrantId(), rerr)
		}
		if herr := hydrateSingleGrant(ctx, c, resolvedSyncID, ret); herr != nil {
			return nil, fmt.Errorf("error hydrating grant '%s': %w", request.GetGrantId(), herr)
		}
	}

	return reader_v2.GrantsReaderServiceGetGrantResponse_builder{
		Grant: ret,
	}.Build(), nil
}

// resolveSyncIDForGrantGet mirrors getConnectorObject's sync_id cascade
// so slim-grant hydration queries the same row.
func (c *C1File) resolveSyncIDForGrantGet(ctx context.Context, explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if c.currentSyncID != "" {
		return c.currentSyncID, nil
	}
	if c.viewSyncID != "" {
		return c.viewSyncID, nil
	}
	latestSyncRun, err := c.getFinishedSync(ctx, 0, connectorstore.SyncTypeAny)
	if err != nil {
		return "", err
	}
	if latestSyncRun == nil {
		latestSyncRun, err = c.getLatestUnfinishedSync(ctx, connectorstore.SyncTypeAny)
		if err != nil {
			return "", err
		}
	}
	if latestSyncRun == nil {
		return "", nil
	}
	return latestSyncRun.ID, nil
}

func (c *C1File) ListGrantsForEntitlement(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForEntitlement")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()
	ret, nextPageToken, err := listGrantsGeneric(ctx, c, request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants for entitlement '%s': %w", request.GetEntitlement().GetId(), err)
	}

	return reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

func (c *C1File) ListGrantsForPrincipal(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForPrincipal")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	ret, nextPageToken, err := listGrantsGeneric(ctx, c, request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants for principal '%s': %w", request.GetPrincipalId(), err)
	}

	return reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

func (c *C1File) ListGrantsForResourceType(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForResourceTypeRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForResourceTypeResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForResourceType")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	ret, nextPageToken, err := listGrantsGeneric(ctx, c, request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants for resource type '%s': %w", request.GetResourceTypeId(), err)
	}

	return reader_v2.GrantsReaderServiceListGrantsForResourceTypeResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

// PutGrants is the connector-facing write method on connectorstore.Writer.
// It replaces any conflicting row and re-extracts expansion metadata from
// the grant payload.
func (c *C1File) PutGrants(ctx context.Context, bulkGrants ...*v2.Grant) error {
	ctx, span := tracer.Start(ctx, "C1File.PutGrants")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	return c.upsertGrants(ctx, grantUpsertOptions{Mode: grantUpsertModeReplace}, bulkGrants...)
}

// PutGrantsIfNewer writes grants only when the provided discovered_at is
// newer than the stored row's discovered_at. Retained on *C1File because
// it has targeted test coverage in grants_test.go documenting its semantics.
// Not on any interface.
func (c *C1File) PutGrantsIfNewer(ctx context.Context, bulkGrants ...*v2.Grant) error {
	ctx, span := tracer.Start(ctx, "C1File.PutGrantsIfNewer")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	return c.upsertGrants(ctx, grantUpsertOptions{Mode: grantUpsertModeIfNewer}, bulkGrants...)
}

// upsertGrants is the internal implementation of grant writes with mode
// dispatch. Exported-surface callers go through PutGrants (Replace),
// PutGrantsIfNewer (IfNewer), or StoreExpandedGrants (PreserveExpansion).
func (c *C1File) upsertGrants(ctx context.Context, opts grantUpsertOptions, bulkGrants ...*v2.Grant) error {
	if c.readOnly {
		return ErrReadOnly
	}
	switch opts.Mode {
	case grantUpsertModeReplace,
		grantUpsertModeIfNewer,
		grantUpsertModePreserveExpansion:
	default:
		return fmt.Errorf("unknown grant upsert mode: %d", opts.Mode)
	}

	if err := upsertGrantsInternal(ctx, c, opts.Mode, bulkGrants...); err != nil {
		return err
	}

	c.dbUpdated = true
	return nil
}

func baseGrantRecord(grant *v2.Grant) goqu.Record {
	return goqu.Record{
		"resource_type_id":           grant.GetEntitlement().GetResource().GetId().GetResourceType(),
		"resource_id":                grant.GetEntitlement().GetResource().GetId().GetResource(),
		"entitlement_id":             grant.GetEntitlement().GetId(),
		"principal_resource_type_id": grant.GetPrincipal().GetId().GetResourceType(),
		"principal_resource_id":      grant.GetPrincipal().GetId().GetResource(),
	}
}

// Hoisted so the per-grant gate doesn't allocate four zero-value
// protos per UpsertGrants call.
var (
	unsafeForSlimSentinels = []proto.Message{
		&v2.InsertResourceGrants{},
		&v2.ExternalResourceMatchAll{},
		&v2.ExternalResourceMatch{},
		&v2.ExternalResourceMatchID{},
	}
)

// unsafeForSlim returns true when a grant's annotations imply the
// syncer reads non-identity fields off its embedded Entitlement.Resource
// or Principal. Slimming silently corrupts those paths.
//
// InsertResourceGrants — the syncer extracts grant.Entitlement.Resource
// and writes it to v1_resources via PutResources. Stubs would overwrite
// the resources table with stripped data on etag-replay.
//
// ExternalResourceMatch{All,Match,ID} — processGrantsWithExternalPrincipals
// builds a bid key from grant.GetPrincipal() that encodes ParentResourceId.
// A slim stub principal has no parent, so its bid misses keys and loses
// transitive expansion through the external-resource match.
func unsafeForSlim(grant *v2.Grant) bool {
	annos := annotations.Annotations(grant.GetAnnotations())
	return annos.ContainsAny(unsafeForSlimSentinels...)
}

func grantExtractFields(c *C1File, mode grantUpsertMode) func(grant *v2.Grant) (goqu.Record, error) {
	return func(grant *v2.Grant) (goqu.Record, error) {
		rec := baseGrantRecord(grant)
		slim := c.v2GrantsWriter && !unsafeForSlim(grant)
		preserveExpansion := mode == grantUpsertModePreserveExpansion

		// PreserveExpansion still must slim the data blob — otherwise
		// expanded grants (the majority for group-heavy tenants) silently
		// stay full-blob.
		if preserveExpansion {
			if !slim {
				return rec, nil
			}
			stripped := proto.Clone(grant).(*v2.Grant)
			slimGrantForWrite(stripped)
			data, err := protoMarshaler.Marshal(stripped)
			if err != nil {
				return nil, fmt.Errorf("error marshaling slim grant (PreserveExpansion): %w", err)
			}
			rec["data"] = data
			return rec, nil
		}

		hasExp := hasGrantExpandable(grant)
		if !hasExp && !slim {
			rec["expansion"] = nil
			rec["needs_expansion"] = false
			return rec, nil
		}

		stripped := proto.Clone(grant).(*v2.Grant)
		var expansionBytes []byte
		var needsExpansion bool
		if hasExp {
			expansionBytes, needsExpansion = extractAndStripExpansion(stripped)
		}
		// Use untyped nil for SQL NULL to avoid driver-specific []byte(nil)->X'' coercion.
		if expansionBytes == nil {
			rec["expansion"] = nil
		} else {
			rec["expansion"] = expansionBytes
		}
		rec["needs_expansion"] = needsExpansion

		if slim {
			slimGrantForWrite(stripped)
		}

		strippedData, err := protoMarshaler.Marshal(stripped)
		if err != nil {
			return nil, fmt.Errorf("error marshaling grant: %w", err)
		}
		rec["data"] = strippedData
		return rec, nil
	}
}

// Id is intentionally kept — stripping it would force GetGrant through
// the slow generic single-row hydration path.
func slimGrantForWrite(grant *v2.Grant) {
	grant.SetEntitlement(nil)
	grant.SetPrincipal(nil)
}

func upsertGrantsInternal(
	ctx context.Context,
	c *C1File,
	mode grantUpsertMode,
	msgs ...*v2.Grant,
) error {
	if len(msgs) == 0 {
		return nil
	}
	ctx, span := tracer.Start(ctx, "C1File.bulkUpsertGrants")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	if err := c.validateSyncDb(ctx); err != nil {
		return err
	}

	rows, err := prepareConnectorObjectRows(c, msgs, grantExtractFields(c, mode))
	if err != nil {
		return err
	}

	return executeGrantChunkedUpsert(ctx, c, rows, mode)
}

// hasGrantExpandable returns true if the grant has a GrantExpandable annotation.
// This is a cheap check (no proto unmarshal) used to avoid cloning grants that
// don't need annotation stripping.
func hasGrantExpandable(grant *v2.Grant) bool {
	expandable := &v2.GrantExpandable{}
	for _, a := range grant.GetAnnotations() {
		if a.MessageIs(expandable) {
			return true
		}
	}
	return false
}

// extractAndStripExpansion extracts the GrantExpandable annotation from the grant,
// removes it from the grant's annotations, and returns the serialized proto bytes.
// The annotation is always stripped from the grant if present, so it never leaks
// into the data blob. Returns (nil, false) if the grant has no GrantExpandable
// annotation or if the annotation contains no valid entitlement IDs.
func extractAndStripExpansion(grant *v2.Grant) ([]byte, bool) {
	annos := annotations.Annotations(grant.GetAnnotations())
	expandable := &v2.GrantExpandable{}
	ok, err := annos.Pick(expandable)
	if err != nil || !ok {
		return nil, false
	}

	// Always strip the GrantExpandable annotation from the grant, regardless
	// of whether it contains valid IDs. This keeps the data blob clean.
	filtered := annotations.Annotations{}
	for _, a := range annos {
		if !a.MessageIs(expandable) {
			filtered = append(filtered, a)
		}
	}
	grant.SetAnnotations(filtered)

	// Only return expansion bytes if there's at least one non-whitespace entitlement ID.
	hasValid := false
	for _, id := range expandable.GetEntitlementIds() {
		if strings.TrimSpace(id) != "" {
			hasValid = true
			break
		}
	}
	if !hasValid {
		return nil, false
	}

	// Serialize the expandable annotation.
	data, err := proto.Marshal(expandable)
	if err != nil {
		return nil, false
	}
	return data, true
}

func backfillGrantExpansionColumn(ctx context.Context, db *goqu.Database, tableName string) error {
	// Backfill grants only for syncs that have not yet been processed.
	// The grants_backfilled flag is the single source of truth for whether
	// this migration work still needs to run for a sync.
	//
	// We unmarshal every grant with expansion IS NULL from old syncs, extract the
	// GrantExpandable annotation (if present), and populate the expansion column.
	// Non-expandable grants get an empty-blob sentinel to avoid re-processing,
	// which is cleaned up to NULL at the end.
	//
	// Uses cursor-based pagination (g.id > ?) so each query jumps to unprocessed
	// rows via the primary key index instead of rescanning from the start.

	// Collect un-backfilled sync IDs upfront. sync_runs is tiny, so this is
	// cheap and lets us skip the grants scan entirely when there's nothing to do.
	syncRows, err := db.QueryContext(ctx, fmt.Sprintf(
		`SELECT sync_id FROM %s WHERE grants_backfilled = 0`, syncRuns.Name(),
	))
	if err != nil {
		return err
	}
	var pendingSyncIDs []any
	for syncRows.Next() {
		var sid string
		if err := syncRows.Scan(&sid); err != nil {
			_ = syncRows.Close()
			return err
		}
		pendingSyncIDs = append(pendingSyncIDs, sid)
	}
	if err := syncRows.Err(); err != nil {
		_ = syncRows.Close()
		return err
	}
	_ = syncRows.Close()

	if len(pendingSyncIDs) == 0 {
		return nil
	}

	placeholders := strings.Repeat("?,", len(pendingSyncIDs))
	placeholders = placeholders[:len(placeholders)-1] // trim trailing comma

	var lastID int64
	for {
		args := make([]any, 0, len(pendingSyncIDs)+1)
		args = append(args, lastID)
		args = append(args, pendingSyncIDs...)

		rows, err := db.QueryContext(ctx, fmt.Sprintf(
			`SELECT g.id, g.data FROM %s g
			 WHERE g.id > ?
			   AND g.expansion IS NULL
			   AND g.sync_id IN (%s)
			 ORDER BY g.id
			 LIMIT 1000`,
			tableName, placeholders,
		), args...)
		if err != nil {
			return err
		}

		type row struct {
			id   int64
			data []byte
		}
		batch := make([]row, 0, 1000)
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.id, &r.data); err != nil {
				_ = rows.Close()
				return err
			}
			batch = append(batch, r)
		}
		if err := rows.Err(); err != nil {
			_ = rows.Close()
			return err
		}
		_ = rows.Close()

		if len(batch) == 0 {
			break
		}

		lastID = batch[len(batch)-1].id

		// Split grants into expandable (need full data rewrite) and
		// non-expandable (just need sentinel marker).
		var sentinelIDs []any
		type expandableRow struct {
			id             int64
			expansionBytes []byte
			needsExpansion bool
			data           []byte
		}
		var expandableRows []expandableRow

		for _, r := range batch {
			g := &v2.Grant{}
			if err := proto.Unmarshal(r.data, g); err != nil {
				return err
			}

			expansionBytes, needsExpansion := extractAndStripExpansion(g)

			if expansionBytes == nil {
				sentinelIDs = append(sentinelIDs, r.id)
				continue
			}

			newData, err := proto.Marshal(g)
			if err != nil {
				return err
			}
			expandableRows = append(expandableRows, expandableRow{
				id:             r.id,
				expansionBytes: expansionBytes,
				needsExpansion: needsExpansion,
				data:           newData,
			})
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		// Batch-mark non-expandable grants with the sentinel in one statement.
		if len(sentinelIDs) > 0 {
			sp := strings.Repeat("?,", len(sentinelIDs))
			sp = sp[:len(sp)-1]
			if _, err := tx.ExecContext(ctx, fmt.Sprintf(
				`UPDATE %s SET expansion=X'' WHERE id IN (%s)`, tableName, sp,
			), sentinelIDs...); err != nil {
				_ = tx.Rollback()
				return err
			}
		}

		// Expandable grants need per-row updates (each has unique data).
		if len(expandableRows) > 0 {
			fullStmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
				`UPDATE %s SET expansion=?, needs_expansion=?, data=? WHERE id=?`, tableName,
			))
			if err != nil {
				_ = tx.Rollback()
				return err
			}
			for _, er := range expandableRows {
				if _, err := fullStmt.ExecContext(ctx, er.expansionBytes, er.needsExpansion, er.data, er.id); err != nil {
					_ = fullStmt.Close()
					_ = tx.Rollback()
					return err
				}
			}
			_ = fullStmt.Close()
		}

		if err := tx.Commit(); err != nil {
			return err
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	// Convert empty-blob sentinels back to NULL.
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET expansion = NULL WHERE expansion = X''`, tableName,
	)); err != nil {
		_ = tx.Rollback()
		return err
	}
	// Mark all not-yet-processed syncs as backfilled so migration work only runs once.
	if _, err := tx.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET grants_backfilled = 1 WHERE grants_backfilled = 0`,
		syncRuns.Name(),
	)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func executeGrantChunkedUpsert(
	ctx context.Context, c *C1File,
	rows []*goqu.Record,
	mode grantUpsertMode,
) error {
	tableName := grants.Name()
	// Expansion column update logic built conditionally in Go so the query planner
	// sees simple expressions instead of parameterized CASE branches.
	var expansionExpr goqu.Expression
	var needsExpansionExpr goqu.Expression

	switch mode {
	case grantUpsertModePreserveExpansion:
		// Keep existing expansion/needs_expansion values on conflict.
		expansionExpr = goqu.L(fmt.Sprintf("%s.expansion", tableName))
		needsExpansionExpr = goqu.L(fmt.Sprintf("%s.needs_expansion", tableName))
	default:
		// Write EXCLUDED expansion/needs_expansion values on conflict.
		// This supports both setting and explicit clearing (expansion=NULL, needs_expansion=0).
		expansionExpr = goqu.L("EXCLUDED.expansion")
		needsExpansionExpr = goqu.L(
			fmt.Sprintf(`CASE
				WHEN EXCLUDED.expansion IS NULL OR EXCLUDED.expansion = X'' THEN 0
				WHEN %[1]s.expansion IS NULL AND EXCLUDED.expansion IS NOT NULL THEN 1
				WHEN %[1]s.expansion IS NOT NULL AND EXCLUDED.expansion IS NOT NULL AND %[1]s.expansion != EXCLUDED.expansion THEN 1
				ELSE %[1]s.needs_expansion
			END`, tableName),
		)
	}

	buildQueryFn := func(insertDs *goqu.InsertDataset, chunkedRows []*goqu.Record) (*goqu.InsertDataset, error) {
		update := goqu.Record{
			"data":            goqu.I("EXCLUDED.data"),
			"expansion":       expansionExpr,
			"needs_expansion": needsExpansionExpr,
		}
		if mode == grantUpsertModeIfNewer {
			update["discovered_at"] = goqu.I("EXCLUDED.discovered_at")
			return insertDs.
				OnConflict(goqu.DoUpdate("external_id, sync_id", update).Where(
					goqu.L("EXCLUDED.discovered_at > ?.discovered_at", goqu.I(tableName)),
				)).
				Rows(chunkedRows).
				Prepared(true), nil
		}

		return insertDs.
			OnConflict(goqu.DoUpdate("external_id, sync_id", update)).
			Rows(chunkedRows).
			Prepared(true), nil
	}

	return executeChunkedInsert(ctx, c, tableName, rows, buildQueryFn)
}

func (c *C1File) DeleteGrant(ctx context.Context, grantId string) error {
	ctx, span := tracer.Start(ctx, "C1File.DeleteGrant")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	err = c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.Delete(grants.Name())
	q = q.Where(goqu.C("external_id").Eq(grantId))
	if c.currentSyncID != "" {
		q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	}
	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
