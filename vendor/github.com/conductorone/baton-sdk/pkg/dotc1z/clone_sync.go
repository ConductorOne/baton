package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

// cloneTableColumns returns the non-autoincrement column names for tableName
// by querying PRAGMA table_info on the given connection. The column names are
// returned in schema-definition order for the source table, which may differ
// from a freshly-created table when columns were added via ALTER TABLE.
func cloneTableColumns(ctx context.Context, conn *sql.Conn, tableName string) ([]string, error) {
	rows, err := conn.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
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

		if err := rows.Scan(&cid, &name, &dataType, &notNull, &defaultValue, &pk); err != nil {
			return nil, err
		}
		if name != "id" {
			columns = append(columns, name)
		}
	}
	return columns, rows.Err()
}

// cloneTableQuery builds an INSERT ... SELECT that copies rows by explicit
// column name rather than relying on SELECT *, which is sensitive to the
// physical column order of the source vs destination tables.
func cloneTableQuery(tableName string, columns []string) string {
	colList := strings.Join(columns, ", ")
	return fmt.Sprintf(
		"INSERT INTO clone.%s (%s) SELECT %s FROM %s WHERE sync_id=?",
		tableName, colList, colList, tableName,
	)
}

// CloneSync uses sqlite hackery to directly copy the pertinent rows into a new database.
// 1. Create a new empty sqlite database in a temp file
// 2. Open the c1z that we are cloning to get a db handle
// 3. Execute an ATTACH query to bring our empty sqlite db into the context of our db connection
// 4. Select directly from the cloned db and insert directly into the new database.
// 5. Close and save the new database as a c1z at the configured path.
func (c *C1File) CloneSync(ctx context.Context, outPath string, syncID string) (err error) {
	ctx, span := tracer.Start(ctx, "C1File.CloneSync")
	defer span.End()

	// Be sure that the output path is empty else return an error
	_, err = os.Stat(outPath)
	if err == nil || !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("clone-sync: output path (%s) must not exist for cloning to proceed", outPath)
	}

	tmpDir, err := os.MkdirTemp(c.tempDir, "c1zclone")
	if err != nil {
		return err
	}

	// Always clean up the temp dir and return an error if that fails
	defer func() {
		cleanupErr := os.RemoveAll(tmpDir)
		if cleanupErr != nil {
			err = errors.Join(err, fmt.Errorf("clone-sync: error cleaning up temp dir: %w", cleanupErr))
		}
	}()

	dbPath := filepath.Join(tmpDir, "db")

	// Create a temporary C1File to initialize the schema in the new db.
	// NewC1File calls init() internally, creating all required tables.
	// We close only the rawDb to release the connection and file locks
	// without triggering C1File.Close()'s cleanupDbDir which would
	// remove the tmpDir we still need.
	initFile, err := NewC1File(ctx, dbPath)
	if err != nil {
		return err
	}
	if err = initFile.rawDb.Close(); err != nil {
		return err
	}
	initFile.rawDb = nil
	initFile.db = nil

	if syncID == "" {
		syncID, err = c.LatestSyncID(ctx, connectorstore.SyncTypeFull)
		if err != nil {
			return err
		}
	}

	sync, err := c.getSync(ctx, syncID)
	if err != nil {
		return err
	}

	if sync == nil {
		return fmt.Errorf("clone-sync: sync not found")
	}

	if sync.EndedAt == nil {
		return fmt.Errorf("clone-sync: sync is not ended")
	}

	qCtx, canc := context.WithCancel(ctx)
	defer canc()

	// Get a single connection to the current db so we can make multiple queries in the same session
	conn, err := c.rawDb.Conn(qCtx)
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = conn.ExecContext(qCtx, fmt.Sprintf(`ATTACH '%s' AS clone`, dbPath))
	if err != nil {
		return err
	}

	for _, t := range allTableDescriptors {
		columns, err := cloneTableColumns(qCtx, conn, t.Name())
		if err != nil {
			return fmt.Errorf("clone-sync: error reading columns for %s: %w", t.Name(), err)
		}
		q := cloneTableQuery(t.Name(), columns)
		_, err = conn.ExecContext(qCtx, q, syncID)
		if err != nil {
			return err
		}
	}

	// Detach the clone database before releasing the connection. On Windows,
	// open file handles prevent deletion; without DETACH the source connection
	// pool retains a handle on the clone db file, causing os.RemoveAll to fail.
	_, err = conn.ExecContext(qCtx, "DETACH clone")
	if err != nil {
		ctxzap.Extract(ctx).Error("error detaching clone database", zap.Error(err))
	}
	canc()
	_ = conn.Close()

	// Open a fresh C1File to compress the populated db into a c1z.
	// No other connections are open on dbPath at this point.
	outFile, err := NewC1File(ctx, dbPath)
	if err != nil {
		return err
	}
	outFile.dbUpdated = true
	outFile.outputFilePath = outPath
	err = outFile.Close(ctx)
	if err != nil {
		return err
	}

	return err
}
