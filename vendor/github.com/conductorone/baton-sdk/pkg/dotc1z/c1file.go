package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	// NOTE: required to register the dialect for goqu.
	//
	// If you remove this import, goqu.Dialect("sqlite3") will
	// return a copy of the default dialect, which is not what we want,
	// and allocates a ton of memory.
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"

	_ "github.com/glebarez/go-sqlite"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
)

var ErrDbNotOpen = errors.New("c1file: database has not been opened")

type pragma struct {
	name  string
	value string
}

type C1File struct {
	rawDb              *sql.DB
	db                 *goqu.Database
	currentSyncID      string
	viewSyncID         string
	outputFilePath     string
	dbFilePath         string
	dbUpdated          bool
	tempDir            string
	pragmas            []pragma
	readOnly           bool
	encoderConcurrency int
	closed             bool
	closedMu           sync.Mutex

	// Cached sync run for listConnectorObjects (avoids N+1 queries)
	cachedViewSyncRun *syncRun
	cachedViewSyncMu  sync.Mutex
	cachedViewSyncErr error

	// Slow query tracking
	slowQueryLogTimes     map[string]time.Time
	slowQueryLogTimesMu   sync.Mutex
	slowQueryThreshold    time.Duration
	slowQueryLogFrequency time.Duration

	// Sync cleanup settings
	syncLimit int
}

var _ connectorstore.InternalWriter = (*C1File)(nil)

type C1FOption func(*C1File)

// WithC1FTmpDir sets the temporary directory to use when cloning a sync.
// If not provided, os.TempDir() will be used.
func WithC1FTmpDir(tempDir string) C1FOption {
	return func(o *C1File) {
		o.tempDir = tempDir
	}
}

// WithC1FPragma sets a sqlite pragma for the c1z file.
func WithC1FPragma(name string, value string) C1FOption {
	return func(o *C1File) {
		o.pragmas = append(o.pragmas, pragma{name, value})
	}
}

func WithC1FReadOnly(readOnly bool) C1FOption {
	return func(o *C1File) {
		o.readOnly = readOnly
	}
}

func WithC1FEncoderConcurrency(concurrency int) C1FOption {
	return func(o *C1File) {
		o.encoderConcurrency = concurrency
	}
}

// WithC1FSyncCountLimit sets the number of syncs to keep during cleanup.
// If not set, defaults to 2 (or BATON_KEEP_SYNC_COUNT env var if set).
func WithC1FSyncCountLimit(limit int) C1FOption {
	return func(o *C1File) {
		o.syncLimit = limit
	}
}

// Returns a C1File instance for the given db filepath.
func NewC1File(ctx context.Context, dbFilePath string, opts ...C1FOption) (*C1File, error) {
	ctx, span := tracer.Start(ctx, "NewC1File")
	defer span.End()

	rawDB, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, fmt.Errorf("new-c1-file: error opening raw db: %w", err)
	}
	l := ctxzap.Extract(ctx)
	l.Debug("new-c1-file: opened raw db",
		zap.String("db_file_path", dbFilePath),
	)

	// Limit to a single connection so idle pool connections don't hold WAL
	// read locks that prevent PRAGMA wal_checkpoint(TRUNCATE) from completing
	// all frames. Without this, saveC1z() can read an incomplete main db file
	// because uncheckpointed WAL frames are invisible to raw file I/O.
	rawDB.SetMaxOpenConns(1)

	db := goqu.New("sqlite3", rawDB)

	c1File := &C1File{
		rawDb:                 rawDB,
		db:                    db,
		dbFilePath:            dbFilePath,
		pragmas:               []pragma{},
		slowQueryLogTimes:     make(map[string]time.Time),
		slowQueryThreshold:    5 * time.Second,
		slowQueryLogFrequency: 1 * time.Minute,
		encoderConcurrency:    1,
	}

	for _, opt := range opts {
		opt(c1File)
	}

	err = c1File.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	err = c1File.init(ctx)
	if err != nil {
		return nil, fmt.Errorf("new-c1-file: error initializing c1file: %w", err)
	}

	return c1File, nil
}

type c1zOptions struct {
	tmpDir             string
	pragmas            []pragma
	decoderOptions     []DecoderOption
	readOnly           bool
	encoderConcurrency int
	syncLimit          int
}

type C1ZOption func(*c1zOptions)

// WithTmpDir sets the temporary directory to extract the c1z file to.
// If not provided, os.TempDir() will be used.
func WithTmpDir(tmpDir string) C1ZOption {
	return func(o *c1zOptions) {
		o.tmpDir = tmpDir
	}
}

// WithPragma sets a sqlite pragma for the c1z file.
func WithPragma(name string, value string) C1ZOption {
	return func(o *c1zOptions) {
		o.pragmas = append(o.pragmas, pragma{name, value})
	}
}

func WithDecoderOptions(opts ...DecoderOption) C1ZOption {
	return func(o *c1zOptions) {
		o.decoderOptions = opts
	}
}

// WithReadOnly opens the c1z file in read only mode. Modifying the c1z will result in an error on close.
func WithReadOnly(readOnly bool) C1ZOption {
	return func(o *c1zOptions) {
		o.readOnly = readOnly
	}
}

// WithEncoderConcurrency sets the number of created encoders.
// Default is 1, which disables async encoding/concurrency.
// 0 uses GOMAXPROCS.
func WithEncoderConcurrency(concurrency int) C1ZOption {
	return func(o *c1zOptions) {
		o.encoderConcurrency = concurrency
	}
}

// WithSyncLimit sets the number of syncs to keep during cleanup.
// If not set, defaults to 2 (or BATON_KEEP_SYNC_COUNT env var if set).
func WithSyncLimit(limit int) C1ZOption {
	return func(o *c1zOptions) {
		o.syncLimit = limit
	}
}

// Returns a new C1File instance with its state stored at the provided filename.
func NewC1ZFile(ctx context.Context, outputFilePath string, opts ...C1ZOption) (*C1File, error) {
	ctx, span := tracer.Start(ctx, "NewC1ZFile")
	defer span.End()

	options := &c1zOptions{
		encoderConcurrency: 1,
	}
	for _, opt := range opts {
		opt(options)
	}
	if options.encoderConcurrency < 0 {
		return nil, fmt.Errorf("encoder concurrency must not be negative: %d", options.encoderConcurrency)
	}

	dbFilePath, _, err := decompressC1z(outputFilePath, options.tmpDir, options.decoderOptions...)
	if err != nil {
		return nil, err
	}
	l := ctxzap.Extract(ctx)
	l.Debug("new-c1z-file: decompressed c1z",
		zap.String("db_file_path", dbFilePath),
		zap.String("output_file_path", outputFilePath),
	)

	var c1fopts []C1FOption
	for _, pragma := range options.pragmas {
		c1fopts = append(c1fopts, WithC1FPragma(pragma.name, pragma.value))
	}
	if options.readOnly {
		c1fopts = append(c1fopts, WithC1FReadOnly(true))
	}
	c1fopts = append(c1fopts, WithC1FEncoderConcurrency(options.encoderConcurrency))
	if options.syncLimit > 0 {
		c1fopts = append(c1fopts, WithC1FSyncCountLimit(options.syncLimit))
	}

	c1File, err := NewC1File(ctx, dbFilePath, c1fopts...)
	if err != nil {
		return nil, cleanupDbDir(dbFilePath, err)
	}

	c1File.outputFilePath = outputFilePath

	return c1File, nil
}

func cleanupDbDir(dbFilePath string, err error) error {
	// Stat dbFilePath to make sure it's a file, not a directory.
	stat, statErr := os.Stat(dbFilePath) //nolint:gosec // G703 -- dbFilePath is a caller-provided path by design.
	if statErr != nil {
		if errors.Is(statErr, os.ErrNotExist) {
			// If the file doesn't exist, we can't clean up the directory.
			return err
		}
		return errors.Join(err, fmt.Errorf("cleanupDbDir: error statting dbFilePath %s: %w", dbFilePath, statErr))
	}
	if stat.IsDir() {
		// If the file is a directory, don't try to clean up the parent directory.
		return errors.Join(err, fmt.Errorf("cleanupDbDir: dbFilePath %s is a directory, not a file: %w", dbFilePath, statErr))
	}

	cleanupErr := os.RemoveAll(filepath.Dir(dbFilePath)) //nolint:gosec // G703 -- dbFilePath is a caller-provided path by design.
	if cleanupErr != nil {
		err = errors.Join(err, cleanupErr)
	}
	return err
}

var ErrReadOnly = errors.New("c1z: read only mode")

// Close ensures that the sqlite database is flushed to disk, and if any changes were made we update the original database
// with our changes. The provided context is used for the WAL checkpoint operation. If the context is already expired,
// a fresh context with a 30-second timeout is used to ensure the checkpoint completes.
func (c *C1File) Close(ctx context.Context) error {
	var err error
	l := ctxzap.Extract(ctx)

	c.closedMu.Lock()
	defer c.closedMu.Unlock()
	if c.closed {
		l.Warn("close called on already-closed c1file", zap.String("db_path", c.dbFilePath))
		return nil
	}

	if c.rawDb != nil {
		// CRITICAL: Force a full WAL checkpoint before closing the database.
		// This ensures all WAL data is written back to the main database file
		// and the writes are synced to disk. Without this, on filesystems with
		// aggressive caching (like ZFS with large ARC), the subsequent saveC1z()
		// read could see stale data because the checkpoint writes may still be
		// in kernel buffers.
		//
		// TRUNCATE mode: checkpoint as many frames as possible, then truncate
		// the WAL file to zero bytes. This guarantees all data is in the main
		// database file before we read it for compression.
		if c.dbUpdated && !c.readOnly {
			// Use a dedicated context for the checkpoint. The caller's context
			// may already be expired (e.g. Temporal activity deadline), but the
			// checkpoint is a local SQLite operation that must complete to avoid
			// saving a stale c1z.
			checkpointCtx := ctx
			if ctx.Err() != nil {
				var checkpointCancel context.CancelFunc
				checkpointCtx, checkpointCancel = context.WithTimeout(context.Background(), 30*time.Second)
				defer checkpointCancel()
			}

			// Use QueryRowContext to read the (busy, log, checkpointed) result.
			// ExecContext silently discards these values, making partial
			// checkpoints undetectable — the PRAGMA returns nil error even when
			// it can't checkpoint all frames due to concurrent readers.
			var busy, log, checkpointed int
			row := c.rawDb.QueryRowContext(checkpointCtx, "PRAGMA wal_checkpoint(TRUNCATE)")
			if err = row.Scan(&busy, &log, &checkpointed); err != nil {
				l.Error("WAL checkpoint failed before close",
					zap.Error(err),
					zap.String("db_path", c.dbFilePath))
				closeErr := c.rawDb.Close()
				if closeErr != nil {
					l.Error("error closing raw db", zap.Error(closeErr))
				}
				c.rawDb = nil
				c.db = nil
				return cleanupDbDir(c.dbFilePath, fmt.Errorf("c1z: WAL checkpoint failed: %w", err))
			}
			if busy != 0 || (log >= 0 && checkpointed < log) {
				l.Error("WAL checkpoint incomplete before close",
					zap.Int("busy", busy),
					zap.Int("log", log),
					zap.Int("checkpointed", checkpointed),
					zap.String("db_path", c.dbFilePath))
				closeErr := c.rawDb.Close()
				if closeErr != nil {
					l.Error("error closing raw db", zap.Error(closeErr))
				}
				c.rawDb = nil
				c.db = nil
				return cleanupDbDir(c.dbFilePath, fmt.Errorf("c1z: WAL checkpoint incomplete: busy=%d log=%d checkpointed=%d", busy, log, checkpointed))
			}
		}

		err = c.rawDb.Close()
		if err != nil {
			return cleanupDbDir(c.dbFilePath, err)
		}
	}
	c.rawDb = nil
	c.db = nil

	// We only want to save the file if we've made any changes
	if c.dbUpdated {
		if c.readOnly {
			return cleanupDbDir(c.dbFilePath, ErrReadOnly)
		}

		// Verify WAL was fully checkpointed. If it still has data,
		// saveC1z would create a c1z missing the WAL contents since
		// it only reads the main database file.
		walPath := c.dbFilePath + "-wal"
		if walInfo, statErr := os.Stat(walPath); statErr == nil && walInfo.Size() > 0 {
			return cleanupDbDir(c.dbFilePath, fmt.Errorf("c1z: WAL file not empty after close (size=%d) - refusing to save incomplete data", walInfo.Size()))
		}

		err = saveC1z(c.dbFilePath, c.outputFilePath, c.encoderConcurrency)
		if err != nil {
			return cleanupDbDir(c.dbFilePath, err)
		}
	}

	err = cleanupDbDir(c.dbFilePath, err)
	if err != nil {
		return err
	}
	c.closed = true

	return nil
}

// init ensures that the database has all of the required schema.
func (c *C1File) init(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "C1File.init")
	defer span.End()

	l := ctxzap.Extract(ctx)

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	err = c.InitTables(ctx)
	if err != nil {
		l.Error("c1file-init: error initializing tables", zap.Error(err))
		return err
	}
	l.Debug("c1file-init: initialized tables",
		zap.String("db_file_path", c.dbFilePath),
	)

	// // Checkpoint the WAL after migrations. Migrations like backfillGrantExpansionColumn
	// // can update many rows, filling the WAL. Without a checkpoint, subsequent reads are
	// // slow because SQLite must scan the WAL hash table for every page read.
	if _, err = c.db.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)"); err != nil {
		l.Warn("c1file-init: WAL checkpoint after init failed", zap.Error(err))
	}

	if c.readOnly {
		// Disable journaling in read only mode, since we're not writing to the database.
		_, err = c.db.ExecContext(ctx, "PRAGMA journal_mode = OFF")
		if err != nil {
			return err
		}
		// Disable synchronous writes in read only mode, since we're not writing to the database.
		_, err = c.db.ExecContext(ctx, "PRAGMA synchronous = OFF")
		if err != nil {
			return err
		}
	}

	hasLockingPragma := false
	for _, pragma := range c.pragmas {
		pragmaName := strings.ToLower(pragma.name)
		if pragmaName == "main.locking_mode" || pragmaName == "locking_mode" {
			hasLockingPragma = true
			break
		}
	}
	if !hasLockingPragma {
		l.Debug("c1file-init: setting locking mode to EXCLUSIVE", zap.String("db_file_path", c.dbFilePath))
		_, err = c.db.ExecContext(ctx, "PRAGMA main.locking_mode = EXCLUSIVE")
		if err != nil {
			return fmt.Errorf("c1file-init: error setting locking mode to EXCLUSIVE: %w", err)
		}
	}

	for _, pragma := range c.pragmas {
		_, err := c.db.ExecContext(ctx, fmt.Sprintf("PRAGMA %s = %s", pragma.name, pragma.value))
		if err != nil {
			return fmt.Errorf("c1file-init: error setting pragma %s = %s: %w", pragma.name, pragma.value, err)
		}
	}

	return nil
}

func (c *C1File) InitTables(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "C1File.InitTables")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	l := ctxzap.Extract(ctx).With(zap.String("db_file_path", c.dbFilePath))
	for _, t := range allTableDescriptors {
		query, args := t.Schema()
		_, err = c.db.ExecContext(ctx, fmt.Sprintf(query, args...))
		if err != nil {
			return fmt.Errorf("c1file-init-tables: error initializing table %s: %w", t.Name(), err)
		}
		l.Debug("c1file-init-tables: initialized table schema, running migrations",
			zap.String("table_name", t.Name()),
		)
		err = t.Migrations(ctx, c.db)
		if err != nil {
			return fmt.Errorf("c1file-init-tables: error running migration for table %s: %w", t.Name(), err)
		}
		l.Debug("c1file-init-tables: ran migrations for table",
			zap.String("table_name", t.Name()),
		)
	}

	return nil
}

// Stats introspects the database and returns the count of objects for the given sync run.
// If syncId is empty, it will use the latest sync run of the given type.
func (c *C1File) Stats(ctx context.Context, syncType connectorstore.SyncType, syncId string) (map[string]int64, error) {
	ctx, span := tracer.Start(ctx, "C1File.Stats")
	defer span.End()

	counts := make(map[string]int64)

	var err error
	if syncId == "" {
		syncId, err = c.LatestSyncID(ctx, syncType)
		if err != nil {
			return nil, err
		}
	}
	resp, err := c.GetSync(ctx, reader_v2.SyncsReaderServiceGetSyncRequest_builder{SyncId: syncId}.Build())
	if err != nil {
		return nil, err
	}
	if resp == nil || !resp.HasSync() {
		return nil, status.Errorf(codes.NotFound, "sync '%s' not found", syncId)
	}
	sync := resp.GetSync()
	if syncType != connectorstore.SyncTypeAny && syncType != connectorstore.SyncType(sync.GetSyncType()) {
		return nil, status.Errorf(codes.InvalidArgument, "sync '%s' is not of type '%s'", syncId, syncType)
	}
	syncType = connectorstore.SyncType(sync.GetSyncType())

	counts["resource_types"] = 0

	var rtStats []*v2.ResourceType
	pageToken := ""
	for {
		resp, err := c.ListResourceTypes(ctx, v2.ResourceTypesServiceListResourceTypesRequest_builder{PageToken: pageToken}.Build())
		if err != nil {
			return nil, err
		}

		rtStats = append(rtStats, resp.GetList()...)

		if resp.GetNextPageToken() == "" {
			break
		}

		pageToken = resp.GetNextPageToken()
	}
	counts["resource_types"] = int64(len(rtStats))
	for _, rt := range rtStats {
		resourceCount, err := c.db.From(resources.Name()).
			Where(goqu.C("resource_type_id").Eq(rt.GetId())).
			Where(goqu.C("sync_id").Eq(syncId)).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}
		counts[rt.GetId()] = resourceCount
	}

	if syncType != connectorstore.SyncTypeResourcesOnly {
		entitlementsCount, err := c.db.From(entitlements.Name()).
			Where(goqu.C("sync_id").Eq(syncId)).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}
		counts["entitlements"] = entitlementsCount

		grantsCount, err := c.db.From(grants.Name()).
			Where(goqu.C("sync_id").Eq(syncId)).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}
		counts["grants"] = grantsCount
	}

	return counts, nil
}

// validateDb ensures that the database has been opened.
func (c *C1File) validateDb(ctx context.Context) error {
	if c.db == nil {
		return ErrDbNotOpen
	}

	return nil
}

// validateSyncDb ensures that there is a sync currently running, and that the database has been opened.
func (c *C1File) validateSyncDb(ctx context.Context) error {
	if c.currentSyncID == "" {
		return fmt.Errorf("c1file: sync is not active")
	}

	return c.validateDb(ctx)
}

func (c *C1File) OutputFilepath() (string, error) {
	if c.outputFilePath == "" {
		return "", fmt.Errorf("c1file: output file path is empty")
	}
	return c.outputFilePath, nil
}

func (c *C1File) AttachFile(other *C1File, dbName string) (*C1FileAttached, error) {
	_, err := c.db.Exec(`ATTACH DATABASE ? AS ?`, other.dbFilePath, dbName)
	if err != nil {
		return nil, err
	}

	return &C1FileAttached{
		safe: true,
		file: c,
	}, nil
}

func (c *C1FileAttached) DetachFile(dbName string) (*C1FileAttached, error) {
	_, err := c.file.db.Exec(`DETACH DATABASE ?`, dbName)
	if err != nil {
		return nil, err
	}

	return &C1FileAttached{
		safe: false,
		file: c.file,
	}, nil
}

// GrantStats introspects the database and returns the count of grants for the given sync run.
// If syncId is empty, it will use the latest sync run of the given type.
func (c *C1File) GrantStats(ctx context.Context, syncType connectorstore.SyncType, syncId string) (map[string]int64, error) {
	ctx, span := tracer.Start(ctx, "C1File.GrantStats")
	defer span.End()

	var err error
	if syncId == "" {
		syncId, err = c.LatestSyncID(ctx, syncType)
		if err != nil {
			return nil, err
		}
	} else {
		lastSync, err := c.GetSync(ctx, reader_v2.SyncsReaderServiceGetSyncRequest_builder{SyncId: syncId}.Build())
		if err != nil {
			return nil, err
		}
		if lastSync == nil {
			return nil, status.Errorf(codes.NotFound, "sync '%s' not found", syncId)
		}
		if syncType != connectorstore.SyncTypeAny && syncType != connectorstore.SyncType(lastSync.GetSync().GetSyncType()) {
			return nil, status.Errorf(codes.InvalidArgument, "sync '%s' is not of type '%s'", syncId, syncType)
		}
	}

	var allResourceTypes []*v2.ResourceType
	pageToken := ""
	for {
		resp, err := c.ListResourceTypes(ctx, v2.ResourceTypesServiceListResourceTypesRequest_builder{PageToken: pageToken}.Build())
		if err != nil {
			return nil, err
		}

		allResourceTypes = append(allResourceTypes, resp.GetList()...)

		if resp.GetNextPageToken() == "" {
			break
		}

		pageToken = resp.GetNextPageToken()
	}

	stats := make(map[string]int64)

	for _, resourceType := range allResourceTypes {
		grantsCount, err := c.db.From(grants.Name()).
			Where(goqu.C("sync_id").Eq(syncId)).
			Where(goqu.C("resource_type_id").Eq(resourceType.GetId())).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}

		stats[resourceType.GetId()] = grantsCount
	}

	return stats, nil
}
