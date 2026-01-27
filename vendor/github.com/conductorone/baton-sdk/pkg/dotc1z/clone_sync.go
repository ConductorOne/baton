package dotc1z

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
)

func cloneTableQuery(tableName string) (string, error) {
	var sb strings.Builder
	var err error

	_, err = sb.WriteString("INSERT INTO clone.")
	if err != nil {
		return "", err
	}

	_, err = sb.WriteString(tableName)
	if err != nil {
		return "", err
	}

	_, err = sb.WriteString(" SELECT * FROM ")
	if err != nil {
		return "", err
	}

	_, err = sb.WriteString(tableName)
	if err != nil {
		return "", err
	}

	_, err = sb.WriteString(" WHERE sync_id=?")
	if err != nil {
		return "", err
	}

	return sb.String(), nil
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
	out, err := NewC1File(ctx, dbPath)
	if err != nil {
		return err
	}
	defer out.Close(ctx)

	err = out.init(ctx)
	if err != nil {
		return err
	}

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
		q, err := cloneTableQuery(t.Name())
		if err != nil {
			return err
		}
		_, err = conn.ExecContext(qCtx, q, syncID)
		if err != nil {
			return err
		}
	}

	// Really be sure that our connection is closed and the db won't be mutated
	canc()
	_ = conn.Close()

	// Hack to wrap the db in a tempdir in a C1Z
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
