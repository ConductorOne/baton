package dotc1z

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
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
	defer span.End()

	ret, nextPageToken, err := listConnectorObjects(ctx, c, grants.Name(), request, func() *v2.Grant { return &v2.Grant{} })
	if err != nil {
		return nil, fmt.Errorf("error listing grants: %w", err)
	}

	return v2.GrantsServiceListGrantsResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

func (c *C1File) GetGrant(ctx context.Context, request *reader_v2.GrantsReaderServiceGetGrantRequest) (*reader_v2.GrantsReaderServiceGetGrantResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.GetGrant")
	defer span.End()

	ret := &v2.Grant{}
	syncId, err := annotations.GetSyncIdFromAnnotations(request.GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error getting sync id from annotations for grant '%s': %w", request.GetGrantId(), err)
	}
	err = c.getConnectorObject(ctx, grants.Name(), request.GetGrantId(), syncId, ret)
	if err != nil {
		return nil, fmt.Errorf("error fetching grant '%s': %w", request.GetGrantId(), err)
	}

	return reader_v2.GrantsReaderServiceGetGrantResponse_builder{
		Grant: ret,
	}.Build(), nil
}

func (c *C1File) ListGrantsForEntitlement(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForEntitlement")
	defer span.End()
	ret, nextPageToken, err := listConnectorObjects(ctx, c, grants.Name(), request, func() *v2.Grant { return &v2.Grant{} })
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
	defer span.End()

	ret, nextPageToken, err := listConnectorObjects(ctx, c, grants.Name(), request, func() *v2.Grant { return &v2.Grant{} })
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
	defer span.End()

	ret, nextPageToken, err := listConnectorObjects(ctx, c, grants.Name(), request, func() *v2.Grant { return &v2.Grant{} })
	if err != nil {
		return nil, fmt.Errorf("error listing grants for resource type '%s': %w", request.GetResourceTypeId(), err)
	}

	return reader_v2.GrantsReaderServiceListGrantsForResourceTypeResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

func (c *C1File) PutGrants(ctx context.Context, bulkGrants ...*v2.Grant) error {
	ctx, span := tracer.Start(ctx, "C1File.PutGrants")
	defer span.End()

	return c.UpsertGrants(ctx, connectorstore.GrantUpsertOptions{
		Mode: connectorstore.GrantUpsertModeReplace,
	}, bulkGrants...)
}

func (c *C1File) PutGrantsIfNewer(ctx context.Context, bulkGrants ...*v2.Grant) error {
	ctx, span := tracer.Start(ctx, "C1File.PutGrantsIfNewer")
	defer span.End()

	return c.UpsertGrants(ctx, connectorstore.GrantUpsertOptions{
		Mode: connectorstore.GrantUpsertModeIfNewer,
	}, bulkGrants...)
}

// UpsertGrants writes grants with explicit conflict semantics.
func (c *C1File) UpsertGrants(ctx context.Context, opts connectorstore.GrantUpsertOptions, bulkGrants ...*v2.Grant) error {
	if c.readOnly {
		return ErrReadOnly
	}
	switch opts.Mode {
	case connectorstore.GrantUpsertModeReplace,
		connectorstore.GrantUpsertModeIfNewer,
		connectorstore.GrantUpsertModePreserveExpansion:
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

func grantExtractFields(mode connectorstore.GrantUpsertMode) func(grant *v2.Grant) (goqu.Record, error) {
	return func(grant *v2.Grant) (goqu.Record, error) {
		rec := baseGrantRecord(grant)
		if mode == connectorstore.GrantUpsertModePreserveExpansion {
			return rec, nil
		}

		if !hasGrantExpandable(grant) {
			rec["expansion"] = nil
			rec["needs_expansion"] = false
			return rec, nil
		}

		stripped := proto.Clone(grant).(*v2.Grant)
		expansionBytes, needsExpansion := extractAndStripExpansion(stripped)
		// Use untyped nil for SQL NULL to avoid driver-specific []byte(nil)->X'' coercion.
		if expansionBytes == nil {
			rec["expansion"] = nil
		} else {
			rec["expansion"] = expansionBytes
		}
		rec["needs_expansion"] = needsExpansion

		strippedData, err := protoMarshaler.Marshal(stripped)
		if err != nil {
			return nil, fmt.Errorf("error marshaling grant after stripping expansion: %w", err)
		}
		rec["data"] = strippedData
		return rec, nil
	}
}

func upsertGrantsInternal(
	ctx context.Context,
	c *C1File,
	mode connectorstore.GrantUpsertMode,
	msgs ...*v2.Grant,
) error {
	if len(msgs) == 0 {
		return nil
	}
	ctx, span := tracer.Start(ctx, "C1File.bulkUpsertGrants")
	defer span.End()

	if err := c.validateSyncDb(ctx); err != nil {
		return err
	}

	rows, err := prepareConnectorObjectRows(c, msgs, grantExtractFields(mode))
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
	// Only backfill grants from syncs that don't support diff (old syncs created before
	// this code change). New syncs set supports_diff=1 at creation and write grants with
	// the expansion column populated correctly, so they don't need backfilling.
	//
	// We unmarshal every grant with expansion IS NULL from old syncs, extract the
	// GrantExpandable annotation (if present), and populate the expansion column.
	// Non-expandable grants get an empty-blob sentinel to avoid re-processing,
	// which is cleaned up to NULL at the end.
	//
	// Uses cursor-based pagination (g.id > ?) so each query jumps to unprocessed
	// rows via the primary key index instead of rescanning from the start.
	var lastID int64
	for {
		rows, err := db.QueryContext(ctx, fmt.Sprintf(
			`SELECT g.id, g.data FROM %s g
			 JOIN %s sr ON g.sync_id = sr.sync_id
			 WHERE g.id > ?
			   AND g.expansion IS NULL
			   AND sr.supports_diff = 0
			 ORDER BY g.id
			 LIMIT 1000`,
			tableName, syncRuns.Name(),
		), lastID)
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

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		stmt, err := tx.PrepareContext(ctx, fmt.Sprintf(
			`UPDATE %s SET expansion=?, needs_expansion=?, data=? WHERE id=?`,
			tableName,
		))
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		for _, r := range batch {
			g := &v2.Grant{}
			if err := proto.Unmarshal(r.data, g); err != nil {
				_ = stmt.Close()
				_ = tx.Rollback()
				return err
			}

			expansionBytes, needsExpansion := extractAndStripExpansion(g)

			// Re-serialize the grant with the annotation stripped.
			newData, err := proto.Marshal(g)
			if err != nil {
				_ = stmt.Close()
				_ = tx.Rollback()
				return err
			}

			// Use empty blob for non-expandable grants so they won't match
			// "expansion IS NULL" on the next iteration. Cleaned up below.
			var expansionVal interface{}
			if expansionBytes != nil {
				expansionVal = expansionBytes
			} else {
				expansionVal = []byte{}
			}

			if _, err := stmt.ExecContext(ctx, expansionVal, needsExpansion, newData, r.id); err != nil {
				_ = stmt.Close()
				_ = tx.Rollback()
				return err
			}
		}

		_ = stmt.Close()
		if err := tx.Commit(); err != nil {
			return err
		}
	}

	// Convert empty-blob sentinels back to NULL.
	if _, err := db.ExecContext(ctx, fmt.Sprintf(
		`UPDATE %s SET expansion = NULL WHERE expansion = X''`, tableName,
	)); err != nil {
		return err
	}

	return nil
}

func executeGrantChunkedUpsert(
	ctx context.Context, c *C1File,
	rows []*goqu.Record,
	mode connectorstore.GrantUpsertMode,
) error {
	tableName := grants.Name()
	// Expansion column update logic built conditionally in Go so the query planner
	// sees simple expressions instead of parameterized CASE branches.
	var expansionExpr goqu.Expression
	var needsExpansionExpr goqu.Expression

	switch mode {
	case connectorstore.GrantUpsertModePreserveExpansion:
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
		if mode == connectorstore.GrantUpsertModeIfNewer {
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
	defer span.End()

	err := c.validateSyncDb(ctx)
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
