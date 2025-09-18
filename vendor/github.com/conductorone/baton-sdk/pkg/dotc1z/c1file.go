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
	rawDb          *sql.DB
	db             *goqu.Database
	currentSyncID  string
	viewSyncID     string
	outputFilePath string
	dbFilePath     string
	dbUpdated      bool
	tempDir        string
	pragmas        []pragma

	// Slow query tracking
	slowQueryLogTimes     map[string]time.Time
	slowQueryLogTimesMu   sync.Mutex
	slowQueryThreshold    time.Duration
	slowQueryLogFrequency time.Duration
}

var _ connectorstore.Writer = (*C1File)(nil)

type C1FOption func(*C1File)

func WithC1FTmpDir(tempDir string) C1FOption {
	return func(o *C1File) {
		o.tempDir = tempDir
	}
}

func WithC1FPragma(name string, value string) C1FOption {
	return func(o *C1File) {
		o.pragmas = append(o.pragmas, pragma{name, value})
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
	tmpDir         string
	pragmas        []pragma
	decoderOptions []DecoderOption
}
type C1ZOption func(*c1zOptions)

func WithTmpDir(tmpDir string) C1ZOption {
	return func(o *c1zOptions) {
		o.tmpDir = tmpDir
	}
}

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

// Returns a new C1File instance with its state stored at the provided filename.
func NewC1ZFile(ctx context.Context, outputFilePath string, opts ...C1ZOption) (*C1File, error) {
	ctx, span := tracer.Start(ctx, "NewC1ZFile")
	defer span.End()

	options := &c1zOptions{}
	for _, opt := range opts {
		opt(options)
	}

	dbFilePath, err := loadC1z(outputFilePath, options.tmpDir, options.decoderOptions...)
	if err != nil {
		return nil, err
	}

	var c1fopts []C1FOption
	for _, pragma := range options.pragmas {
		c1fopts = append(c1fopts, WithC1FPragma(pragma.name, pragma.value))
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

// Close ensures that the sqlite database is flushed to disk, and if any changes were made we update the original database
// with our changes.
func (c *C1File) Close() error {
	var err error

	if c.rawDb != nil {
		err = c.rawDb.Close()
		if err != nil {
			return cleanupDbDir(c.dbFilePath, err)
		}
	}
	c.rawDb = nil
	c.db = nil

	// We only want to save the file if we've made any changes
	if c.dbUpdated {
		err = saveC1z(c.dbFilePath, c.outputFilePath)
		if err != nil {
			return cleanupDbDir(c.dbFilePath, err)
		}
	}

	return cleanupDbDir(c.dbFilePath, err)
}

// init ensures that the database has all of the required schema.
func (c *C1File) init(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "C1File.init")
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

	for _, pragma := range c.pragmas {
		_, err := c.db.ExecContext(ctx, fmt.Sprintf("PRAGMA %s = %s", pragma.name, pragma.value))
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
	resp, err := c.GetSync(ctx, &reader_v2.SyncsReaderServiceGetSyncRequest{SyncId: syncId})
	if err != nil {
		return nil, err
	}
	if resp == nil || resp.Sync == nil {
		return nil, status.Errorf(codes.NotFound, "sync '%s' not found", syncId)
	}
	sync := resp.Sync
	if syncType != connectorstore.SyncTypeAny && syncType != connectorstore.SyncType(sync.SyncType) {
		return nil, status.Errorf(codes.InvalidArgument, "sync '%s' is not of type '%s'", syncId, syncType)
	}
	syncType = connectorstore.SyncType(sync.SyncType)

	counts["resource_types"] = 0

	var rtStats []*v2.ResourceType
	pageToken := ""
	for {
		resp, err := c.ListResourceTypes(ctx, &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken})
		if err != nil {
			return nil, err
		}

		rtStats = append(rtStats, resp.List...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}
	counts["resource_types"] = int64(len(rtStats))
	for _, rt := range rtStats {
		resourceCount, err := c.db.From(resources.Name()).
			Where(goqu.C("resource_type_id").Eq(rt.Id)).
			Where(goqu.C("sync_id").Eq(syncId)).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}
		counts[rt.Id] = resourceCount
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
		lastSync, err := c.GetSync(ctx, &reader_v2.SyncsReaderServiceGetSyncRequest{SyncId: syncId})
		if err != nil {
			return nil, err
		}
		if lastSync == nil {
			return nil, status.Errorf(codes.NotFound, "sync '%s' not found", syncId)
		}
		if syncType != connectorstore.SyncTypeAny && syncType != connectorstore.SyncType(lastSync.Sync.SyncType) {
			return nil, status.Errorf(codes.InvalidArgument, "sync '%s' is not of type '%s'", syncId, syncType)
		}
	}

	var allResourceTypes []*v2.ResourceType
	pageToken := ""
	for {
		resp, err := c.ListResourceTypes(ctx, &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken})
		if err != nil {
			return nil, err
		}

		allResourceTypes = append(allResourceTypes, resp.List...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	stats := make(map[string]int64)

	for _, resourceType := range allResourceTypes {
		grantsCount, err := c.db.From(grants.Name()).
			Where(goqu.C("sync_id").Eq(syncId)).
			Where(goqu.C("resource_type_id").Eq(resourceType.Id)).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}

		stats[resourceType.Id] = grantsCount
	}

	return stats, nil
}
