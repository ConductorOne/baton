package dotc1z

import (
	"context"
	"errors"
	"fmt"
)

type C1FileAttached struct {
	safe bool
	file *C1File
}

func (c *C1FileAttached) CompactTable(ctx context.Context, destSyncID string, baseSyncID string, appliedSyncID string, tableName string) error {
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

	// Step 1: Insert ALL records from base sync
	insertBaseQuery := fmt.Sprintf(`
		INSERT INTO main.%s (%s)
		SELECT %s
		FROM base.%s
		WHERE sync_id = ?
	`, tableName, columnList, selectList, tableName)

	_, err = c.file.db.ExecContext(ctx, insertBaseQuery, destSyncID, baseSyncID)
	if err != nil {
		return fmt.Errorf("failed to copy base records: %w", err)
	}

	// Step 2: Insert/replace records from applied sync where applied.discovered_at > main.discovered_at
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

	_, err = c.file.db.ExecContext(ctx, insertOrReplaceAppliedQuery, destSyncID, appliedSyncID, destSyncID, destSyncID)
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
		var defaultValue interface{}

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

func (c *C1FileAttached) CompactResourceTypes(ctx context.Context, destSyncID string, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, destSyncID, baseSyncID, appliedSyncID, "v1_resource_types")
}

func (c *C1FileAttached) CompactResources(ctx context.Context, destSyncID string, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, destSyncID, baseSyncID, appliedSyncID, "v1_resources")
}

func (c *C1FileAttached) CompactEntitlements(ctx context.Context, destSyncID string, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, destSyncID, baseSyncID, appliedSyncID, "v1_entitlements")
}

func (c *C1FileAttached) CompactGrants(ctx context.Context, destSyncID string, baseSyncID string, appliedSyncID string) error {
	if !c.safe {
		return errors.New("database has been detached")
	}
	return c.CompactTable(ctx, destSyncID, baseSyncID, appliedSyncID, "v1_grants")
}
