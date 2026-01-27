package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

var _ connectorstore.Writer = (*C1File)(nil)

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
		return nil, err
	}

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
		return nil, err
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

	dbFilePath, _, err := decompressC1z(outputFilePath, options.tmpDir, options.decoderOptions...)
	if err != nil {
		return nil, err
	}

	var c1fopts []C1FOption
	for _, pragma := range options.pragmas {
		c1fopts = append(c1fopts, WithC1FPragma(pragma.name, pragma.value))
	}
	if options.readOnly {
		c1fopts = append(c1fopts, WithC1FReadOnly(true))
	}
	if options.encoderConcurrency < 0 {
		return nil, fmt.Errorf("encoder concurrency must be greater than 0")
	}
	c1fopts = append(c1fopts, WithC1FEncoderConcurrency(options.encoderConcurrency))
	if options.syncLimit > 0 {
		c1fopts = append(c1fopts, WithC1FSyncCountLimit(options.syncLimit))
	}

	c1File, err := NewC1File(ctx, dbFilePath, c1fopts...)
	if err != nil {
		return nil, err
	}

	c1File.outputFilePath = outputFilePath

	return c1File, nil
}

func cleanupDbDir(dbFilePath string, err error) error {
	cleanupErr := os.RemoveAll(filepath.Dir(dbFilePath))
	if cleanupErr != nil {
		err = errors.Join(err, cleanupErr)
	}
	return err
}

var ErrReadOnly = errors.New("c1z: read only mode")

// Close ensures that the sqlite database is flushed to disk, and if any changes were made we update the original database
// with our changes. The provided context is used for the WAL checkpoint operation.
func (c *C1File) Close(ctx context.Context) error {
	var err error

	c.closedMu.Lock()
	defer c.closedMu.Unlock()
	if c.closed {
		l := ctxzap.Extract(ctx)
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
			_, err = c.rawDb.ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
			if err != nil {
				l := ctxzap.Extract(ctx)
				// Checkpoint failed - log and continue. The subsequent Close()
				// will attempt a passive checkpoint. If that also fails, we'll
				// get an error from Close() or saveC1z() will read stale data.
				// We log here for debugging but don't fail because:
				// 1. Close() will still attempt its own checkpoint
				// 2. The error might be transient (busy)
				l.Warn("WAL checkpoint failed before close",
					zap.Error(err),
					zap.String("db_path", c.dbFilePath))
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

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	err = c.InitTables(ctx)
	if err != nil {
		return err
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

	for _, pragma := range c.pragmas {
		_, err := c.db.ExecContext(ctx, fmt.Sprintf("PRAGMA %s = %s", pragma.name, pragma.value))
		if err != nil {
			return err
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

	for _, t := range allTableDescriptors {
		query, args := t.Schema()
		_, err = c.db.ExecContext(ctx, fmt.Sprintf(query, args...))
		if err != nil {
			return err
		}
		err = t.Migrations(ctx, c.db)
		if err != nil {
			return err
		}
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
		return fmt.Errorf("c1file: datbase has not been opened")
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
