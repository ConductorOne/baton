package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
)

const syncRunsTableVersion = "1"
const syncRunsTableName = "sync_runs"
const syncRunsTableSchema = `
create table if not exists %s (
    id integer primary key,
    sync_id text not null,
    started_at datetime not null,
    ended_at datetime,
    sync_token text not null,
    sync_type text not null default 'full',
    parent_sync_id text not null default ''
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

func (r *syncRunsTable) Migrations(ctx context.Context, db *goqu.Database) error {
	// Check if sync_type column exists
	var syncTypeExists int
	err := db.QueryRowContext(ctx, fmt.Sprintf("select count(*) from pragma_table_info('%s') where name='sync_type'", r.Name())).Scan(&syncTypeExists)
	if err != nil {
		return err
	}
	if syncTypeExists == 0 {
		_, err = db.ExecContext(ctx, fmt.Sprintf("alter table %s add column sync_type text not null default 'full'", r.Name()))
		if err != nil {
			return err
		}
	}

	// Check if parent_sync_id column exists
	var parentSyncIDExists int
	err = db.QueryRowContext(ctx, fmt.Sprintf("select count(*) from pragma_table_info('%s') where name='parent_sync_id'", r.Name())).Scan(&parentSyncIDExists)
	if err != nil {
		return err
	}
	if parentSyncIDExists == 0 {
		_, err = db.ExecContext(ctx, fmt.Sprintf("alter table %s add column parent_sync_id text not null default ''", r.Name()))
		if err != nil {
			return err
		}
	}

	return nil
}

type syncRun struct {
	ID           string
	StartedAt    *time.Time
	EndedAt      *time.Time
	SyncToken    string
	Type         connectorstore.SyncType
	ParentSyncID string
}

func (c *C1File) getLatestUnfinishedSync(ctx context.Context, syncType connectorstore.SyncType) (*syncRun, error) {
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
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token", "sync_type", "parent_sync_id")
	q = q.Where(goqu.C("ended_at").IsNull())
	q = q.Where(goqu.C("started_at").Gte(oneWeekAgo))
	q = q.Order(goqu.C("started_at").Desc())
	if syncType != connectorstore.SyncTypeAny {
		q = q.Where(goqu.C("sync_type").Eq(syncType))
	}
	q = q.Limit(1)

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, err
	}

	row := c.db.QueryRowContext(ctx, query, args...)

	err = row.Scan(&ret.ID, &ret.StartedAt, &ret.EndedAt, &ret.SyncToken, &ret.Type, &ret.ParentSyncID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return ret, nil
}

func (c *C1File) getFinishedSync(ctx context.Context, offset uint, syncType connectorstore.SyncType) (*syncRun, error) {
	ctx, span := tracer.Start(ctx, "C1File.getFinishedSync")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return nil, err
	}

	// Validate syncType
	if !slices.Contains(connectorstore.AllSyncTypes, syncType) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid sync type: %s", syncType)
	}

	ret := &syncRun{}
	q := c.db.From(syncRuns.Name())
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token", "sync_type", "parent_sync_id")
	q = q.Where(goqu.C("ended_at").IsNotNull())
	if syncType != connectorstore.SyncTypeAny {
		q = q.Where(goqu.C("sync_type").Eq(syncType))
	}
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

	err = row.Scan(&ret.ID, &ret.StartedAt, &ret.EndedAt, &ret.SyncToken, &ret.Type, &ret.ParentSyncID)
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
	q = q.Select("id", "sync_id", "started_at", "ended_at", "sync_token", "sync_type", "parent_sync_id")

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
		err := rows.Scan(&rowId, &data.ID, &data.StartedAt, &data.EndedAt, &data.SyncToken, &data.Type, &data.ParentSyncID)
		if err != nil {
			return nil, "", err
		}
		lastRow = rowId
		ret = append(ret, data)
	}
	if rows.Err() != nil {
		return nil, "", rows.Err()
	}

	nextPageToken := ""
	if count > pageSize {
		nextPageToken = strconv.Itoa(lastRow + 1)
	}

	return ret, nextPageToken, nil
}

func (c *C1File) LatestSyncID(ctx context.Context, syncType connectorstore.SyncType) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.LatestSyncID")
	defer span.End()

	s, err := c.getFinishedSync(ctx, 0, syncType)
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

func (c *C1File) PreviousSyncID(ctx context.Context, syncType connectorstore.SyncType) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.PreviousSyncID")
	defer span.End()

	s, err := c.getFinishedSync(ctx, 1, syncType)
	if err != nil {
		return "", err
	}

	if s == nil {
		return "", nil
	}

	return s.ID, nil
}

func (c *C1File) LatestFinishedSyncID(ctx context.Context, syncType connectorstore.SyncType) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.LatestFinishedSync")
	defer span.End()

	s, err := c.getFinishedSync(ctx, 0, syncType)
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
	q = q.Select("sync_id", "started_at", "ended_at", "sync_token", "sync_type", "parent_sync_id")
	q = q.Where(goqu.C("sync_id").Eq(syncID))

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, err
	}
	row := c.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&ret.ID, &ret.StartedAt, &ret.EndedAt, &ret.SyncToken, &ret.Type, &ret.ParentSyncID)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *C1File) getCurrentSync(ctx context.Context) (*syncRun, error) {
	ctx, span := tracer.Start(ctx, "C1File.getCurrentSync")
	defer span.End()

	if c.currentSyncID == "" {
		return nil, fmt.Errorf("c1file: sync must be running to get current sync")
	}

	return c.getSync(ctx, c.currentSyncID)
}

func (c *C1File) SetCurrentSync(ctx context.Context, syncID string) error {
	ctx, span := tracer.Start(ctx, "C1File.SetCurrentSync")
	defer span.End()

	_, err := c.getSync(ctx, syncID)
	if err != nil {
		return err
	}

	c.currentSyncID = syncID
	return nil
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

func (c *C1File) ResumeSync(ctx context.Context, syncType connectorstore.SyncType, syncID string) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.ResumeSync")
	defer span.End()

	if c.currentSyncID != "" {
		if syncID == c.currentSyncID {
			return c.currentSyncID, nil
		}
		if syncID != "" {
			return "", status.Errorf(codes.FailedPrecondition, "current sync is %s, cannot resume %s", c.currentSyncID, syncID)
		}
	}

	if syncID != "" {
		syncRun, err := c.getSync(ctx, syncID)
		if err != nil {
			return "", err
		}
		if syncType != connectorstore.SyncTypeAny && syncRun.Type != syncType {
			return "", status.Errorf(codes.FailedPrecondition, "cannot resume sync (%s) when a different sync type (%s) is running", syncRun.Type, syncType)
		}
		if syncRun.EndedAt != nil {
			return "", status.Errorf(codes.FailedPrecondition, "cannot resume sync that has already ended")
		}
		c.currentSyncID = syncID
		return c.currentSyncID, nil
	}

	if c.currentSyncID != "" {
		syncRun, err := c.getSync(ctx, c.currentSyncID)
		if err != nil {
			return "", err
		}
		if syncType != connectorstore.SyncTypeAny && syncRun.Type != syncType {
			return "", status.Errorf(codes.FailedPrecondition, "cannot resume sync. current sync %s is type %s, cannot resume as type %s", syncRun.ID, syncRun.Type, syncType)
		}
		if syncRun.EndedAt != nil {
			return "", status.Errorf(codes.Internal, "current sync %s has already ended. this should never happen", syncRun.ID)
		}

		return c.currentSyncID, nil
	}

	syncRun, err := c.getLatestUnfinishedSync(ctx, syncType)
	if err != nil {
		return "", err
	}
	if syncRun == nil {
		return "", status.Errorf(codes.NotFound, "no unfinished sync found for type %s", syncType)
	}

	c.currentSyncID = syncRun.ID
	return c.currentSyncID, nil
}

// StartOrResumeSync checks if a sync is already running and resumes it if it is.
// If no sync is running, it starts a new sync.
// It returns the sync ID and a boolean indicating if a new sync was started.
func (c *C1File) StartOrResumeSync(ctx context.Context, syncType connectorstore.SyncType, syncID string) (string, bool, error) {
	ctx, span := tracer.Start(ctx, "C1File.StartOrResumeSync")
	defer span.End()

	resumedSyncID, err := c.ResumeSync(ctx, syncType, syncID)
	if err != nil {
		if status.Code(err) != codes.NotFound && !errors.Is(err, sql.ErrNoRows) {
			return "", false, err
		}
	} else {
		return resumedSyncID, false, nil
	}

	if syncID != "" {
		return "", false, status.Errorf(codes.NotFound, "no sync with id %s found to resume", syncID)
	}

	syncID, err = c.StartNewSync(ctx, syncType, "")
	if err != nil {
		return "", false, err
	}

	c.currentSyncID = syncID

	return c.currentSyncID, true, nil
}

func (c *C1File) StartNewSync(ctx context.Context, syncType connectorstore.SyncType, parentSyncID string) (string, error) {
	ctx, span := tracer.Start(ctx, "C1File.StartNewSync")
	defer span.End()

	if c.currentSyncID != "" {
		cur, err := c.getSync(ctx, c.currentSyncID)
		if err != nil {
			return "", err
		}
		if cur != nil && cur.EndedAt == nil && cur.Type != syncType {
			return "", status.Errorf(codes.FailedPrecondition, "current sync (id %s) is type %s. cannot start %s", cur.ID, cur.Type, syncType)
		}
		return c.currentSyncID, nil
	}

	switch syncType {
	case connectorstore.SyncTypeFull:
		if parentSyncID != "" {
			return "", status.Errorf(codes.InvalidArgument, "parent sync id must be empty for full sync")
		}
	case connectorstore.SyncTypeResourcesOnly:
		if parentSyncID != "" {
			return "", status.Errorf(codes.InvalidArgument, "parent sync id must be empty for resources only sync")
		}
	case connectorstore.SyncTypePartial:
	case connectorstore.SyncTypeAny:
		return "", status.Errorf(codes.InvalidArgument, "sync cannot be started with SyncTypeAny")
	default:
		return "", status.Errorf(codes.InvalidArgument, "invalid sync type: %s", syncType)
	}

	syncID := ksuid.New().String()

	if err := c.insertSyncRun(ctx, syncID, syncType, parentSyncID); err != nil {
		return "", err
	}

	c.currentSyncID = syncID

	return c.currentSyncID, nil
}

func (c *C1File) insertSyncRun(ctx context.Context, syncID string, syncType connectorstore.SyncType, parentSyncID string) error {
	q := c.db.Insert(syncRuns.Name())
	q = q.Rows(goqu.Record{
		"sync_id":        syncID,
		"started_at":     time.Now().Format("2006-01-02 15:04:05.999999999"),
		"sync_token":     "",
		"sync_type":      syncType,
		"parent_sync_id": parentSyncID,
	})

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

	if err := c.endSyncRun(ctx, c.currentSyncID); err != nil {
		return err
	}

	c.currentSyncID = ""

	return nil
}

func (c *C1File) endSyncRun(ctx context.Context, syncID string) error {
	q := c.db.Update(syncRuns.Name())
	q = q.Set(goqu.Record{
		"ended_at": time.Now().Format("2006-01-02 15:04:05.999999999"),
	})
	q = q.Where(goqu.C("sync_id").Eq(syncID))
	q = q.Where(goqu.C("ended_at").IsNull())

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
	var partials []*syncRun

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
			if sr.Type == connectorstore.SyncTypePartial || sr.Type == connectorstore.SyncTypeResourcesOnly {
				partials = append(partials, sr)
			} else {
				ret = append(ret, sr)
			}
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

	// Delete non-full syncs that ended before the earliest-kept full sync started
	if len(ret) > syncLimit {
		earliestKeptSync := ret[len(ret)-syncLimit]
		l.Debug("Earliest kept sync", zap.String("sync_id", earliestKeptSync.ID), zap.Time("started_at", *earliestKeptSync.StartedAt))

		for _, partial := range partials {
			if partial.EndedAt != nil && partial.EndedAt.Before(*earliestKeptSync.StartedAt) {
				err = c.DeleteSyncRun(ctx, partial.ID)
				if err != nil {
					return err
				}
				l.Info("Removed partial sync that ended before earliest kept sync.",
					zap.String("partial_sync_end", partial.EndedAt.Format(time.RFC3339)),
					zap.String("earliest_kept_sync_start", earliestKeptSync.StartedAt.Format(time.RFC3339)),
					zap.String("sync_id", partial.ID))
			}
		}
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
			Id:           sr.ID,
			StartedAt:    toTimeStamp(sr.StartedAt),
			EndedAt:      toTimeStamp(sr.EndedAt),
			SyncToken:    sr.SyncToken,
			SyncType:     string(sr.Type),
			ParentSyncId: sr.ParentSyncID,
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
			Id:           sr.ID,
			StartedAt:    toTimeStamp(sr.StartedAt),
			EndedAt:      toTimeStamp(sr.EndedAt),
			SyncToken:    sr.SyncToken,
			SyncType:     string(sr.Type),
			ParentSyncId: sr.ParentSyncID,
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

	sync, err := c.getFinishedSync(ctx, 0, connectorstore.SyncType(request.SyncType))
	if err != nil {
		return nil, fmt.Errorf("error fetching latest finished sync: %w", err)
	}

	if sync == nil {
		return &reader_v2.SyncsReaderServiceGetLatestFinishedSyncResponse{
			Sync: nil,
		}, nil
	}

	return &reader_v2.SyncsReaderServiceGetLatestFinishedSyncResponse{
		Sync: &reader_v2.SyncRun{
			Id:           sync.ID,
			StartedAt:    toTimeStamp(sync.StartedAt),
			EndedAt:      toTimeStamp(sync.EndedAt),
			SyncToken:    sync.SyncToken,
			SyncType:     string(sync.Type),
			ParentSyncId: sync.ParentSyncID,
		},
	}, nil
}
