package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/doug-martin/goqu/v9"
	"github.com/segmentio/ksuid"
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
		if col == "sync_id" { //nolint:goconst,nolintlint // ...
			selectList += "? as sync_id" //nolint:goconst,nolintlint // ...
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

// GenerateSyncDiffFromFile compares the old sync (in attached) with the new sync (in main)
// and generates two new syncs in the main database.
//
// IMPORTANT: This assumes main=NEW/compacted and attached=OLD/base:
// - diffTableFromAttached: items in attached (OLD) not in main (NEW) = deletions
// - diffTableFromMain: items in main (NEW) not in attached (OLD) = upserts (additions)
//
// Parameters:
// - oldSyncID: the sync ID in the attached database (OLD/base state)
// - newSyncID: the sync ID in the main database (NEW/compacted state)
//
// Returns (upsertsSyncID, deletionsSyncID, error).
func (c *C1FileAttached) GenerateSyncDiffFromFile(ctx context.Context, oldSyncID string, newSyncID string) (string, string, error) {
	if !c.safe {
		return "", "", errors.New("database has been detached")
	}

	ctx, span := tracer.Start(ctx, "C1FileAttached.GenerateSyncDiffFromFile")
	defer span.End()

	// Generate unique IDs for the diff syncs
	deletionsSyncID := ksuid.New().String()
	upsertsSyncID := ksuid.New().String()

	// Start transaction for atomicity
	tx, err := c.file.rawDb.BeginTx(ctx, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure rollback on error
	committed := false
	defer func() {
		if !committed {
			_ = tx.Rollback()
		}
	}()

	now := time.Now().Format("2006-01-02 15:04:05.999999999")

	// Create the deletions sync first (so upserts is "latest")
	// Link it to upserts sync bidirectionally
	deletionsInsert := c.file.db.Insert(syncRuns.Name()).Rows(goqu.Record{
		"sync_id":        deletionsSyncID,
		"started_at":     now,
		"sync_token":     "",
		"sync_type":      connectorstore.SyncTypePartialDeletions,
		"parent_sync_id": oldSyncID,
		"linked_sync_id": upsertsSyncID,
	})
	query, args, err := deletionsInsert.ToSQL()
	if err != nil {
		return "", "", fmt.Errorf("failed to build deletions sync insert: %w", err)
	}
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return "", "", fmt.Errorf("failed to create deletions sync: %w", err)
	}

	// Create the upserts sync, linked to deletions sync
	upsertsInsert := c.file.db.Insert(syncRuns.Name()).Rows(goqu.Record{
		"sync_id":        upsertsSyncID,
		"started_at":     now,
		"sync_token":     "",
		"sync_type":      connectorstore.SyncTypePartialUpserts,
		"parent_sync_id": oldSyncID,
		"linked_sync_id": deletionsSyncID,
	})
	query, args, err = upsertsInsert.ToSQL()
	if err != nil {
		return "", "", fmt.Errorf("failed to build upserts sync insert: %w", err)
	}
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return "", "", fmt.Errorf("failed to create upserts sync: %w", err)
	}

	// Process each table
	// main=NEW, attached=OLD
	// - diffTableFromAttachedTx finds items in OLD not in NEW = deletions
	// - diffTableFromMainTx finds items in NEW not in OLD or modified = upserts
	tables := []string{"v1_resource_types", "v1_resources", "v1_entitlements", "v1_grants"}
	for _, tableName := range tables {
		if err := c.diffTableFromAttachedTx(ctx, tx, tableName, oldSyncID, newSyncID, deletionsSyncID); err != nil {
			return "", "", fmt.Errorf("failed to generate deletions for %s: %w", tableName, err)
		}
		if err := c.diffTableFromMainTx(ctx, tx, tableName, oldSyncID, newSyncID, upsertsSyncID); err != nil {
			return "", "", fmt.Errorf("failed to generate upserts for %s: %w", tableName, err)
		}
	}

	// End the syncs (deletions first, then upserts)
	endedAt := time.Now().Format("2006-01-02 15:04:05.999999999")

	endDeletions := c.file.db.Update(syncRuns.Name()).
		Set(goqu.Record{"ended_at": endedAt}).
		Where(goqu.C("sync_id").Eq(deletionsSyncID), goqu.C("ended_at").IsNull())
	query, args, err = endDeletions.ToSQL()
	if err != nil {
		return "", "", fmt.Errorf("failed to build end deletions sync: %w", err)
	}
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return "", "", fmt.Errorf("failed to end deletions sync: %w", err)
	}

	endUpserts := c.file.db.Update(syncRuns.Name()).
		Set(goqu.Record{"ended_at": endedAt}).
		Where(goqu.C("sync_id").Eq(upsertsSyncID), goqu.C("ended_at").IsNull())
	query, args, err = endUpserts.ToSQL()
	if err != nil {
		return "", "", fmt.Errorf("failed to build end upserts sync: %w", err)
	}
	if _, err = tx.ExecContext(ctx, query, args...); err != nil {
		return "", "", fmt.Errorf("failed to end upserts sync: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return "", "", fmt.Errorf("failed to commit transaction: %w", err)
	}
	committed = true
	c.file.dbUpdated = true

	return upsertsSyncID, deletionsSyncID, nil
}

// diffTableFromAttachedTx finds items in attached (OLD) that don't exist in main (NEW).
// These are DELETIONS - items that existed before but no longer exist.
// Uses the provided transaction.
func (c *C1FileAttached) diffTableFromAttachedTx(ctx context.Context, tx *sql.Tx, tableName string, oldSyncID string, newSyncID string, targetSyncID string) error {
	columns, err := c.getTableColumns(ctx, tableName)
	if err != nil {
		return err
	}

	// Build column lists
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

	// Insert items from attached (OLD) that don't exist in main (NEW)
	// oldSyncID is in attached, newSyncID is in main
	//nolint:gosec // table names are from hardcoded list, not user input
	query := fmt.Sprintf(`
		INSERT INTO main.%s (%s)
		SELECT %s
		FROM attached.%s AS a
		WHERE a.sync_id = ?
		  AND NOT EXISTS (
		    SELECT 1 FROM main.%s AS m 
		    WHERE m.external_id = a.external_id AND m.sync_id = ?
		  )
	`, tableName, columnList, selectList, tableName, tableName)

	_, err = tx.ExecContext(ctx, query, targetSyncID, oldSyncID, newSyncID)
	return err
}

// diffTableFromMainTx finds items in main (NEW) that are new or modified compared to attached (OLD).
// These are UPSERTS - items that are new or have changed.
// Uses the provided transaction.
func (c *C1FileAttached) diffTableFromMainTx(ctx context.Context, tx *sql.Tx, tableName string, oldSyncID string, newSyncID string, targetSyncID string) error {
	columns, err := c.getTableColumns(ctx, tableName)
	if err != nil {
		return err
	}

	// Build column lists
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

	// Insert items from main (NEW) that are:
	// 1. Not in attached (OLD) - additions
	// 2. In attached but with different data - modifications
	// newSyncID is in main, oldSyncID is in attached
	//nolint:gosec // table names are from hardcoded list, not user input
	query := fmt.Sprintf(`
		INSERT INTO main.%s (%s)
		SELECT %s
		FROM main.%s AS m
		WHERE m.sync_id = ?
		  AND (
		    NOT EXISTS (
		      SELECT 1 FROM attached.%s AS a 
		      WHERE a.external_id = m.external_id AND a.sync_id = ?
		    )
		    OR EXISTS (
		      SELECT 1 FROM attached.%s AS a 
		      WHERE a.external_id = m.external_id 
		        AND a.sync_id = ?
		        AND a.data != m.data
		    )
		  )
	`, tableName, columnList, selectList, tableName, tableName, tableName)

	_, err = tx.ExecContext(ctx, query, targetSyncID, newSyncID, oldSyncID, oldSyncID)
	return err
}
