package dotc1z

import (
	"context"
	"errors"
	"fmt"

	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/doug-martin/goqu/v9"
)

type C1FileAttached struct {
	safe bool
	file *C1File
}

func (c *C1FileAttached) CompactTable(ctx context.Context, baseSyncID string, appliedSyncID string, tableName string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	ctx, span := tracer.Start(ctx, "C1FileAttached.CompactTable")
	defer span.End()

	// Get the column structure for this table by querying the schema
	columns, err := c.getTableColumns(ctx, tableName)
	if err != nil {
		return fmt.Errorf("failed to get table columns: %w", err)
	}

	// Build column lists for INSERT statements
	columnList := ""
	selectList := ""
	for i, col := range columns {
		if i > 0 {
			columnList += ", "
			selectList += ", "
		}
		columnList += col
		if col == "sync_id" {
			selectList += "? as sync_id"
		} else {
			selectList += col
		}
	}

	// Insert/replace records from applied sync where applied.discovered_at > main.discovered_at
	insertOrReplaceAppliedQuery := fmt.Sprintf(`
		INSERT OR REPLACE INTO main.%s (%s)
		SELECT %s
		FROM attached.%s AS a
		WHERE a.sync_id = ?
		  AND (
		    NOT EXISTS (
		      SELECT 1 FROM main.%s AS m 
		      WHERE m.external_id = a.external_id AND m.sync_id = ?
		    )
		    OR EXISTS (
		      SELECT 1 FROM main.%s AS m 
		      WHERE m.external_id = a.external_id 
		        AND m.sync_id = ?
		        AND a.discovered_at > m.discovered_at
		    )
		  )
	`, tableName, columnList, selectList, tableName, tableName, tableName)

	_, err = c.file.db.ExecContext(ctx, insertOrReplaceAppliedQuery, baseSyncID, appliedSyncID, baseSyncID, baseSyncID)
	return err
}

func (c *C1FileAttached) getTableColumns(ctx context.Context, tableName string) ([]string, error) {
	if !c.safe {
		return nil, errors.New("database has been detached")
	}
	// PRAGMA doesn't support parameter binding, so we format the table name directly
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := c.file.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull, pk int
		var defaultValue any

		err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk)
		if err != nil {
			return nil, err
		}

		// Skip the 'id' column as it's auto-increment
		if name != "id" {
			columns = append(columns, name)
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return columns, nil
}

func (c *C1FileAttached) CompactResourceTypes(ctx context.Context, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, baseSyncID, appliedSyncID, "v1_resource_types")
}

func (c *C1FileAttached) CompactResources(ctx context.Context, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, baseSyncID, appliedSyncID, "v1_resources")
}

func (c *C1FileAttached) CompactEntitlements(ctx context.Context, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, baseSyncID, appliedSyncID, "v1_entitlements")
}

func (c *C1FileAttached) CompactGrants(ctx context.Context, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, baseSyncID, appliedSyncID, "v1_grants")
}

func unionSyncTypes(a, b connectorstore.SyncType) connectorstore.SyncType {
	switch {
	case a == connectorstore.SyncTypeFull || b == connectorstore.SyncTypeFull:
		return connectorstore.SyncTypeFull
	case a == connectorstore.SyncTypeResourcesOnly || b == connectorstore.SyncTypeResourcesOnly:
		return connectorstore.SyncTypeResourcesOnly
	default:
		return connectorstore.SyncTypePartial
	}
}

func (c *C1FileAttached) UpdateSync(ctx context.Context, baseSync *reader_v2.SyncRun, appliedSync *reader_v2.SyncRun) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	syncType := unionSyncTypes(connectorstore.SyncType(baseSync.GetSyncType()), connectorstore.SyncType(appliedSync.GetSyncType()))

	latestEndedAt := baseSync.GetEndedAt().AsTime()
	if appliedSync.GetEndedAt().AsTime().After(latestEndedAt) {
		latestEndedAt = appliedSync.GetEndedAt().AsTime()
	}

	baseSyncID := baseSync.GetId()
	q := c.file.db.Update(fmt.Sprintf("main.%s", syncRuns.Name()))
	q = q.Set(goqu.Record{
		"ended_at":  latestEndedAt.Format("2006-01-02 15:04:05.999999999"),
		"sync_type": string(syncType),
	})
	q = q.Where(goqu.C("sync_id").Eq(baseSyncID))

	query, args, err := q.ToSQL()
	if err != nil {
		return fmt.Errorf("failed to build update sync query: %w", err)
	}

	_, err = c.file.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update sync %s to type %s: %w", baseSyncID, syncType, err)
	}

	return nil
}
