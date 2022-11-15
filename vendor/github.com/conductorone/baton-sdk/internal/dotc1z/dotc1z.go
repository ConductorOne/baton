package dotc1z

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/glebarez/go-sqlite"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

type C1File struct {
	rawDb          *sql.DB
	db             *goqu.Database
	currentSyncID  string
	viewSyncID     string
	outputFilePath string
	dbFilePath     string
	dbUpdated      bool
}

// Returns a C1File instance for the given db filepath.
func NewC1File(ctx context.Context, dbFilePath string) (*C1File, error) {
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

	err = c1File.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	return c1File, nil
}

// Returns a new C1File instance with its state stored at the provided filename.
func NewC1ZFile(ctx context.Context, outputFilePath string) (*C1File, error) {
	dbFilePath, err := loadC1z(outputFilePath)
	if err != nil {
		return nil, err
	}

	c1File, err := NewC1File(ctx, dbFilePath)
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
	if c.rawDb != nil {
		err := c.rawDb.Close()
		if err != nil {
			return err
		}
	}
	c.rawDb = nil
	c.db = nil

	// We only want to save the file if we've made any changes
	if c.dbUpdated {
		err := saveC1z(c.dbFilePath, c.outputFilePath)
		if err != nil {
			return err
		}
	}

	return nil
}

// init ensures that the database has all of the required schema.
func (c *C1File) init(ctx context.Context) error {
	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	tables := []tableDescriptor{
		resourceTypes,
		resources,
		entitlements,
		grants,
		syncRuns,
		assets,
	}

	for _, t := range tables {
		query, args := t.Schema()

		_, err = c.db.ExecContext(ctx, fmt.Sprintf(query, args...))
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
