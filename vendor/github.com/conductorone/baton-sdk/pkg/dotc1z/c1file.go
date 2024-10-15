package dotc1z

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/doug-martin/goqu/v9"
	// NOTE: required to register the dialect for goqu.
	//
	// If you remove this import, goqu.Dialect("sqlite3") will
	// return a copy of the default dialect, which is not what we want,
	// and allocates a ton of memory.
	_ "github.com/doug-martin/goqu/v9/dialect/sqlite3"

	_ "github.com/glebarez/go-sqlite"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
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
	rawDB, err := sql.Open("sqlite", dbFilePath)
	if err != nil {
		return nil, err
	}

	db := goqu.New("sqlite3", rawDB)

	c1File := &C1File{
		rawDb:      rawDB,
		db:         db,
		dbFilePath: dbFilePath,
	}

	for _, opt := range opts {
		opt(c1File)
	}

	err = c1File.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	return c1File, nil
}

type c1zOptions struct {
	tmpDir  string
	pragmas []pragma
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

// Returns a new C1File instance with its state stored at the provided filename.
func NewC1ZFile(ctx context.Context, outputFilePath string, opts ...C1ZOption) (*C1File, error) {
	options := &c1zOptions{}
	for _, opt := range opts {
		opt(options)
	}

	dbFilePath, err := loadC1z(outputFilePath, options.tmpDir)
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

	err = c1File.init(ctx)
	if err != nil {
		return nil, err
	}

	return c1File, nil
}

// Close ensures that the sqlite database is flushed to disk, and if any changes were made we update the original database
// with our changes.
func (c *C1File) Close() error {
	var err error

	if c.rawDb != nil {
		err = c.rawDb.Close()
		if err != nil {
			return err
		}
	}
	c.rawDb = nil
	c.db = nil

	// We only want to save the file if we've made any changes
	if c.dbUpdated {
		err = saveC1z(c.dbFilePath, c.outputFilePath)
		if err != nil {
			return err
		}
	}

	// Cleanup the database filepath. This should always be a file within a temp directory, so we remove the entire dir.
	err = os.RemoveAll(filepath.Dir(c.dbFilePath))
	if err != nil {
		return err
	}

	return nil
}

// init ensures that the database has all of the required schema.
func (c *C1File) init(ctx context.Context) error {
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
func (c *C1File) Stats(ctx context.Context) (map[string]int64, error) {
	counts := make(map[string]int64)

	syncID, err := c.LatestSyncID(ctx)
	if err != nil {
		return nil, err
	}

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
			Where(goqu.C("sync_id").Eq(syncID)).
			CountContext(ctx)
		if err != nil {
			return nil, err
		}
		counts[rt.Id] = resourceCount
	}

	entitlementsCount, err := c.db.From(entitlements.Name()).
		Where(goqu.C("sync_id").Eq(syncID)).
		CountContext(ctx)
	if err != nil {
		return nil, err
	}
	counts["entitlements"] = entitlementsCount

	grantsCount, err := c.db.From(grants.Name()).
		Where(goqu.C("sync_id").Eq(syncID)).
		CountContext(ctx)
	if err != nil {
		return nil, err
	}

	counts["grants"] = grantsCount

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
