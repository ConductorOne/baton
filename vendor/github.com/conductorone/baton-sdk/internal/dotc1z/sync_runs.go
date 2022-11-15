package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/segmentio/ksuid"
)

const syncRunsTableVersion = "1"
const syncRunsTableName = "sync_runs"
const syncRunsTableSchema = `
create table if not exists %s (
    id integer primary key,
    sync_id text not null,
    started_at datetime not null,
    ended_at datetime,
    sync_token text not null
);
create unique index if not exists %s on %s (sync_id);`

var syncRuns = (*syncRunsTable)(nil)

type syncRunsTable struct{}

func (r *syncRunsTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), syncRunsTableName)
}

func (r *syncRunsTable) Version() string {
	return syncRunsTableVersion
}

func (r *syncRunsTable) Schema() (string, []interface{}) {
	return syncRunsTableSchema, []interface{}{
		r.Name(),
		fmt.Sprintf("idx_sync_runs_sync_id_v%s", r.Version()),
		r.Name(),
	}
}

type syncRun struct {
	ID        string
	StartedAt *time.Time
	EndedAt   *time.Time
	SyncToken string
}

func (c *C1File) getLatestUnfinishedSync(ctx context.Context) (*syncRun, error) {
	err := c.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	ret := &syncRun{}
	q := c.db.From(syncRuns.Name())
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token")
	q = q.Where(goqu.C("ended_at").IsNull())
	q = q.Order(goqu.C("started_at").Desc())
	q = q.Limit(1)

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	row := c.db.QueryRowContext(ctx, query, args...)

	err = row.Scan(&ret.ID, &ret.StartedAt, &ret.EndedAt, &ret.SyncToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return ret, nil
}

func (c *C1File) getFinishedSync(ctx context.Context, offset uint) (*syncRun, error) {
	err := c.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	ret := &syncRun{}
	q := c.db.From(syncRuns.Name())
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token")
	q = q.Where(goqu.C("ended_at").IsNotNull())
	q = q.Order(goqu.C("ended_at").Desc())
	q = q.Limit(1)

	if offset != 0 {
		q = q.Offset(offset)
	}

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	row := c.db.QueryRowContext(ctx, query, args...)

	err = row.Scan(&ret.ID, &ret.StartedAt, &ret.EndedAt, &ret.SyncToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return ret, nil
}

func (c *C1File) LatestSyncID(ctx context.Context) (string, error) {
	s, err := c.getFinishedSync(ctx, 0)
	if err != nil {
		return "", err
	}

	if s == nil {
		return "", nil
	}

	return s.ID, nil
}

func (c *C1File) ViewSync(ctx context.Context, syncID string) error {
	if c.currentSyncID != "" {
		return fmt.Errorf("cannot set view when sync is running")
	}

	c.viewSyncID = syncID

	return nil
}

func (c *C1File) PreviousSyncID(ctx context.Context) (string, error) {
	s, err := c.getFinishedSync(ctx, 1)
	if err != nil {
		return "", err
	}

	if s == nil {
		return "", nil
	}

	return s.ID, nil
}

func (c *C1File) getSync(ctx context.Context, syncID string) (*syncRun, error) {
	err := c.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	ret := &syncRun{}

	q := c.db.From(syncRuns.Name())
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token")
	q = q.Where(goqu.C("sync_id").Eq(syncID))

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	row := c.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&ret.ID, &ret.StartedAt, &ret.EndedAt, &ret.SyncToken)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *C1File) getCurrentSync(ctx context.Context) (*syncRun, error) {
	if c.currentSyncID == "" {
		return nil, fmt.Errorf("c1file: sync must be running to checkpoint")
	}

	return c.getSync(ctx, c.currentSyncID)
}

func (c *C1File) CheckpointSync(ctx context.Context, syncToken string) error {
	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.Update(syncRuns.Name())
	q = q.Set(goqu.Record{"sync_token": syncToken})
	q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))

	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	c.dbUpdated = true

	return nil
}

// StartSync generates a sync ID to be associated with all objects discovered during this run.
func (c *C1File) StartSync(ctx context.Context) (string, bool, error) {
	if c.currentSyncID != "" {
		return c.currentSyncID, false, nil
	}

	newSync := false

	sync, err := c.getLatestUnfinishedSync(ctx)
	if err != nil {
		return "", false, err
	}

	syncID := ksuid.New().String()
	if sync != nil && sync.EndedAt == nil {
		syncID = sync.ID
	} else {
		q := c.db.Insert(syncRuns.Name())
		q = q.Rows(goqu.Record{
			"sync_id":    syncID,
			"started_at": time.Now().Format("2006-01-02 15:04:05.999999999"),
			"sync_token": "",
		})

		query, args, err := q.ToSQL()
		if err != nil {
			return "", false, err
		}

		_, err = c.db.ExecContext(ctx, query, args...)
		if err != nil {
			return "", false, err
		}

		newSync = true
		c.dbUpdated = true
	}

	c.currentSyncID = syncID

	return c.currentSyncID, newSync, nil
}

func (c *C1File) CurrentSyncStep(ctx context.Context) (string, error) {
	sr, err := c.getCurrentSync(ctx)
	if err != nil {
		return "", err
	}

	return sr.SyncToken, nil
}

// EndSync updates the current sync_run row with the end time, and removes any other objects that don't have the current sync ID.
func (c *C1File) EndSync(ctx context.Context) error {
	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.Update(syncRuns.Name())
	q = q.Set(goqu.Record{
		"ended_at": time.Now().Format("2006-01-02 15:04:05.999999999"),
	})
	q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	q = q.Where(goqu.C("ended_at").IsNull())

	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	c.currentSyncID = ""
	c.dbUpdated = true

	return nil
}
