package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"

	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
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
	ctx, span := tracer.Start(ctx, "C1File.getLatestUnfinishedSync")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	// Don't resume syncs that started over a week ago
	oneWeekAgo := time.Now().AddDate(0, 0, -7)
	ret := &syncRun{}
	q := c.db.From(syncRuns.Name())
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token")
	q = q.Where(goqu.C("ended_at").IsNull())
	q = q.Where(goqu.C("started_at").Gte(oneWeekAgo))
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
	ctx, span := tracer.Start(ctx, "C1File.getFinishedSync")
	defer span.End()

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

func (c *C1File) ListSyncRuns(ctx context.Context, pageToken string, pageSize uint32) ([]*syncRun, string, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListSyncRuns")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return nil, "", err
	}

	q := c.db.From(syncRuns.Name()).Prepared(true)
	q = q.Select("id", "sync_id", "started_at", "ended_at", "sync_token")

	if pageToken != "" {
		q = q.Where(goqu.C("id").Gte(pageToken))
	}

	if pageSize > maxPageSize || pageSize <= 0 {
		pageSize = maxPageSize
	}

	q = q.Order(goqu.C("id").Asc())
	q = q.Limit(uint(pageSize + 1))

	var ret []*syncRun

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, "", err
	}

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var count uint32 = 0
	lastRow := 0
	for rows.Next() {
		count++
		if count > pageSize {
			break
		}
		rowId := 0
		data := &syncRun{}
		err := rows.Scan(&rowId, &data.ID, &data.StartedAt, &data.EndedAt, &data.SyncToken)
		if err != nil {
			return nil, "", err
		}
		lastRow = rowId
		ret = append(ret, data)
	}

	nextPageToken := ""
	if count > pageSize {
		nextPageToken = strconv.Itoa(lastRow + 1)
	}

	return ret, nextPageToken, nil
}

func (c *C1File) LatestSyncID(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.LatestSyncID")
	defer span.End()

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
	ctx, span := tracer.Start(ctx, "C1File.PreviousSyncID")
	defer span.End()

	s, err := c.getFinishedSync(ctx, 1)
	if err != nil {
		return "", err
	}

	if s == nil {
		return "", nil
	}

	return s.ID, nil
}

func (c *C1File) LatestFinishedSync(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.LatestFinishedSync")
	defer span.End()

	s, err := c.getFinishedSync(ctx, 0)
	if err != nil {
		return "", err
	}

	if s == nil {
		return "", nil
	}

	return s.ID, nil
}

func (c *C1File) getSync(ctx context.Context, syncID string) (*syncRun, error) {
	ctx, span := tracer.Start(ctx, "C1File.getSync")
	defer span.End()

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
	ctx, span := tracer.Start(ctx, "C1File.getCurrentSync")
	defer span.End()

	if c.currentSyncID == "" {
		return nil, fmt.Errorf("c1file: sync must be running to checkpoint")
	}

	return c.getSync(ctx, c.currentSyncID)
}

func (c *C1File) CheckpointSync(ctx context.Context, syncToken string) error {
	ctx, span := tracer.Start(ctx, "C1File.CheckpointSync")
	defer span.End()

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
	ctx, span := tracer.Start(ctx, "C1File.StartSync")
	defer span.End()

	if c.currentSyncID != "" {
		return c.currentSyncID, false, nil
	}

	newSync := false

	sync, err := c.getLatestUnfinishedSync(ctx)
	if err != nil {
		return "", false, err
	}

	var syncID string
	if sync != nil && sync.EndedAt == nil {
		syncID = sync.ID
	} else {
		syncID, err = c.StartNewSync(ctx)
		if err != nil {
			return "", false, err
		}
		newSync = true
	}

	c.currentSyncID = syncID

	return c.currentSyncID, newSync, nil
}

func (c *C1File) StartNewSync(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.StartNewSync")
	defer span.End()

	// Not sure if we want to do this here
	if c.currentSyncID != "" {
		return c.currentSyncID, nil
	}

	syncID := ksuid.New().String()

	q := c.db.Insert(syncRuns.Name())
	q = q.Rows(goqu.Record{
		"sync_id":    syncID,
		"started_at": time.Now().Format("2006-01-02 15:04:05.999999999"),
		"sync_token": "",
	})

	query, args, err := q.ToSQL()
	if err != nil {
		return "", err
	}

	_, err = c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return "", err
	}

	c.dbUpdated = true
	c.currentSyncID = syncID

	return c.currentSyncID, nil
}

func (c *C1File) CurrentSyncStep(ctx context.Context) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.CurrentSyncStep")
	defer span.End()

	sr, err := c.getCurrentSync(ctx)
	if err != nil {
		return "", err
	}

	return sr.SyncToken, nil
}

// EndSync updates the current sync_run row with the end time, and removes any other objects that don't have the current sync ID.
func (c *C1File) EndSync(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "C1File.EndSync")
	defer span.End()

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

func (c *C1File) Cleanup(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "C1File.Cleanup")
	defer span.End()

	l := ctxzap.Extract(ctx)

	if skipCleanup, _ := strconv.ParseBool(os.Getenv("BATON_SKIP_CLEANUP")); skipCleanup {
		l.Info("BATON_SKIP_CLEANUP is set, skipping cleanup of old syncs")
		return nil
	}

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	if c.currentSyncID != "" {
		l.Warn("current sync is running, skipping cleanup of old syncs", zap.String("current_sync_id", c.currentSyncID))
		return nil
	}

	var ret []*syncRun

	pageToken := ""
	for {
		runs, nextPageToken, err := c.ListSyncRuns(ctx, pageToken, 100)
		if err != nil {
			return err
		}

		for _, sr := range runs {
			if sr.EndedAt == nil {
				continue
			}
			ret = append(ret, sr)
		}

		if nextPageToken == "" {
			break
		}
		pageToken = nextPageToken
	}

	syncLimit := 2
	if customSyncLimit, err := strconv.ParseInt(os.Getenv("BATON_KEEP_SYNC_COUNT"), 10, 64); err == nil && customSyncLimit > 0 {
		syncLimit = int(customSyncLimit)
	}

	l.Debug("found syncs", zap.Int("count", len(ret)), zap.Int("sync_limit", syncLimit))
	if len(ret) <= syncLimit {
		return nil
	}

	l.Info("Cleaning up old sync data...")
	for i := 0; i < len(ret)-syncLimit; i++ {
		err = c.DeleteSyncRun(ctx, ret[i].ID)
		if err != nil {
			return err
		}
		l.Info("Removed old sync data.", zap.String("sync_date", ret[i].EndedAt.Format(time.RFC3339)), zap.String("sync_id", ret[i].ID))
	}

	err = c.Vacuum(ctx)
	if err != nil {
		return err
	}

	c.dbUpdated = true

	return nil
}

// DeleteSyncRun removes all the objects with a given syncID from the database.
func (c *C1File) DeleteSyncRun(ctx context.Context, syncID string) error {
	ctx, span := tracer.Start(ctx, "C1File.DeleteSyncRun")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	// Bail if we're actively syncing
	if c.currentSyncID != "" && c.currentSyncID == syncID {
		return fmt.Errorf("unable to delete the current active sync run")
	}

	for _, t := range allTableDescriptors {
		q := c.db.Delete(t.Name())
		q = q.Where(goqu.C("sync_id").Eq(syncID))

		query, args, err := q.ToSQL()
		if err != nil {
			return err
		}

		_, err = c.db.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}
	}
	c.dbUpdated = true

	return nil
}

// Vacuum runs a VACUUM on the database to reclaim space.
func (c *C1File) Vacuum(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "C1File.Vacuum")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	_, err = c.rawDb.ExecContext(ctx, "VACUUM")
	if err != nil {
		return err
	}

	c.dbUpdated = true

	return nil
}

func toTimeStamp(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return timestamppb.New(*t)
}

func (c *C1File) GetSync(ctx context.Context, request *reader_v2.SyncsReaderServiceGetSyncRequest) (*reader_v2.SyncsReaderServiceGetSyncResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.GetSync")
	defer span.End()

	sr, err := c.getSync(ctx, request.SyncId)
	if err != nil {
		return nil, fmt.Errorf("error getting sync '%s': %w", request.SyncId, err)
	}

	return &reader_v2.SyncsReaderServiceGetSyncResponse{
		Sync: &reader_v2.SyncRun{
			Id:        sr.ID,
			StartedAt: toTimeStamp(sr.StartedAt),
			EndedAt:   toTimeStamp(sr.EndedAt),
			SyncToken: sr.SyncToken,
		},
	}, nil
}

func (c *C1File) ListSyncs(ctx context.Context, request *reader_v2.SyncsReaderServiceListSyncsRequest) (*reader_v2.SyncsReaderServiceListSyncsResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListSyncs")
	defer span.End()

	syncs, nextPageToken, err := c.ListSyncRuns(ctx, request.PageToken, request.PageSize)
	if err != nil {
		return nil, fmt.Errorf("error listing syncs: %w", err)
	}

	syncRuns := make([]*reader_v2.SyncRun, len(syncs))
	for i, sr := range syncs {
		syncRuns[i] = &reader_v2.SyncRun{
			Id:        sr.ID,
			StartedAt: toTimeStamp(sr.StartedAt),
			EndedAt:   toTimeStamp(sr.EndedAt),
			SyncToken: sr.SyncToken,
		}
	}

	return &reader_v2.SyncsReaderServiceListSyncsResponse{
		Syncs:         syncRuns,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) GetLatestFinishedSync(ctx context.Context, request *reader_v2.SyncsReaderServiceGetLatestFinishedSyncRequest) (*reader_v2.SyncsReaderServiceGetLatestFinishedSyncResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.GetLatestFinishedSync")
	defer span.End()

	sync, err := c.getFinishedSync(ctx, 0)
	if err != nil {
		return nil, fmt.Errorf("error fetching latest finished sync: %w", err)
	}

	return &reader_v2.SyncsReaderServiceGetLatestFinishedSyncResponse{
		Sync: &reader_v2.SyncRun{
			Id:        sync.ID,
			StartedAt: toTimeStamp(sync.StartedAt),
			EndedAt:   toTimeStamp(sync.EndedAt),
			SyncToken: sync.SyncToken,
		},
	}, nil
}
