package dotc1z

import (
	"context"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/segmentio/ksuid"
)

func (c *C1File) GenerateSyncDiff(ctx context.Context, baseSyncID string, appliedSyncID string) (string, error) {
	// Validate that both sync runs exist
	baseSync, err := c.getSync(ctx, baseSyncID)
	if err != nil {
		return "", err
	}
	if baseSync == nil {
		return "", fmt.Errorf("generate-diff: base sync not found")
	}

	newSync, err := c.getSync(ctx, appliedSyncID)
	if err != nil {
		return "", err
	}
	if newSync == nil {
		return "", fmt.Errorf("generate-diff: new sync not found")
	}

	// Generate a new unique ID for the diff sync
	diffSyncID := ksuid.New().String()

	if err := c.insertSyncRun(ctx, diffSyncID, SyncTypePartial, baseSyncID); err != nil {
		return "", err
	}

	for _, t := range allTableDescriptors {
		if strings.Contains(t.Name(), syncRunsTableName) {
			continue
		}

		q, args, err := c.diffTableQuery(t, baseSyncID, appliedSyncID, diffSyncID)
		if err != nil {
			return "", err
		}
		_, err = c.db.ExecContext(ctx, q, args...)
		if err != nil {
			return "", err
		}
		c.dbUpdated = true
	}

	if err := c.endSyncRun(ctx, diffSyncID); err != nil {
		return "", err
	}

	return diffSyncID, nil
}

func (c *C1File) diffTableQuery(table tableDescriptor, baseSyncID, appliedSyncID, newSyncID string) (string, []any, error) {
	// Define the columns to select based on the table name
	columns := []interface{}{
		"external_id",
		"data",
		"sync_id",
		"discovered_at",
	}

	tableName := table.Name()
	// Add table-specific columns
	switch {
	case strings.Contains(tableName, resourcesTableName):
		columns = append(columns, "resource_type_id", "parent_resource_type_id", "parent_resource_id")
	case strings.Contains(tableName, resourceTypesTableName):
		// Nothing new to add here
	case strings.Contains(tableName, grantsTableName):
		columns = append(columns, "resource_type_id", "resource_id", "entitlement_id", "principal_resource_type_id", "principal_resource_id")
	case strings.Contains(tableName, entitlementsTableName):
		columns = append(columns, "resource_type_id", "resource_id")
	case strings.Contains(tableName, assetsTableName):
		columns = append(columns, "content_type")
	}

	// Build the subquery to find external_ids in the base sync
	subquery := c.db.Select("external_id").
		From(tableName).
		Where(goqu.C("sync_id").Eq(baseSyncID))

	queryColumns := []interface{}{}
	for _, col := range columns {
		if col == "sync_id" {
			queryColumns = append(queryColumns, goqu.L(fmt.Sprintf("'%s' as sync_id", newSyncID)))
			continue
		}
		queryColumns = append(queryColumns, col)
	}

	// Build the main query to select records from newSyncID that don't exist in baseSyncID
	query := c.db.Insert(tableName).
		Cols(columns...).
		Prepared(true).
		FromQuery(
			c.db.Select(queryColumns...).
				From(tableName).
				Where(
					goqu.C("sync_id").Eq(appliedSyncID),
					goqu.C("external_id").NotIn(subquery),
				),
		)

	// Generate the SQL and args
	return query.ToSQL()
}
