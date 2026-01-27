package dotc1z

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

const bulkPutParallelThreshold = 100
const insertChunkSize = 200
const maxPageSize = 10000

// Use worker pool to limit goroutines.
var numWorkers = min(max(runtime.GOMAXPROCS(0), 1), 4)

var allTableDescriptors = []tableDescriptor{
	resourceTypes,
	resources,
	entitlements,
	grants,
	syncRuns,
	assets,
	sessionStore,
}

type tableDescriptor interface {
	Name() string
	Schema() (string, []any)
	Version() string
	Migrations(ctx context.Context, db *goqu.Database) error
}

type listRequest interface {
	proto.Message
	GetPageSize() uint32
	GetPageToken() string
	GetAnnotations() []*anypb.Any
}

type hasResourceTypeListRequest interface {
	listRequest
	GetResourceTypeId() string
}

type hasResourceIdListRequest interface {
	listRequest
	GetResourceId() *v2.ResourceId
}

type hasResourceListRequest interface {
	listRequest
	GetResource() *v2.Resource
}

type hasEntitlementListRequest interface {
	listRequest
	GetEntitlement() *v2.Entitlement
}

type hasPrincipalIdListRequest interface {
	listRequest
	GetPrincipalId() *v2.ResourceId
}

type hasPrincipalResourceTypeIDsListRequest interface {
	listRequest
	GetPrincipalResourceTypeIds() []string
}

type protoHasID interface {
	proto.Message
	GetId() string
}

// throttledWarnSlowQuery logs a warning about a slow query at most once per minute per request type.
func (c *C1File) throttledWarnSlowQuery(ctx context.Context, query string, duration time.Duration) {
	c.slowQueryLogTimesMu.Lock()
	defer c.slowQueryLogTimesMu.Unlock()

	now := time.Now()
	lastLogTime, exists := c.slowQueryLogTimes[query]
	if !exists || now.Sub(lastLogTime) > c.slowQueryLogFrequency {
		ctxzap.Extract(ctx).Warn(
			"slow query detected",
			zap.String("query", query),
			zap.Duration("duration", duration),
		)
		c.slowQueryLogTimes[query] = now
	}
}

// listConnectorObjects uses a connector list request to fetch the corresponding data from the local db.
// It returns a slice of typed proto messages constructed via the provided factory function.
func listConnectorObjects[T proto.Message](ctx context.Context, c *C1File, tableName string, req listRequest, factory func() T) ([]T, string, error) {
	ctx, span := tracer.Start(ctx, "C1File.listConnectorObjects")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return nil, "", err
	}

	annoSyncID, err := annotations.GetSyncIdFromAnnotations(req.GetAnnotations())
	if err != nil {
		return nil, "", fmt.Errorf("error getting sync id from annotations for list request: %w", err)
	}

	var reqSyncID string
	switch {
	// If the request has a sync id annotation, use that
	case annoSyncID != "":
		reqSyncID = annoSyncID

	// We are currently syncing, so use the current sync id
	case c.currentSyncID != "":
		reqSyncID = c.currentSyncID

	// We are viewing a sync, so use the view sync id
	case c.viewSyncID != "":
		reqSyncID = c.viewSyncID

	// Be explicit that we have no sync ID set
	default:
		reqSyncID = ""
	}

	q := c.db.From(tableName).Prepared(true)
	q = q.Select("id", "data")

	// If the request allows filtering by resource type, apply the filter
	if resourceTypeReq, ok := req.(hasResourceTypeListRequest); ok {
		rt := resourceTypeReq.GetResourceTypeId()
		if rt != "" {
			q = q.Where(goqu.C("resource_type_id").Eq(rt))
		}
	}

	if resourceIdReq, ok := req.(hasResourceIdListRequest); ok {
		r := resourceIdReq.GetResourceId()
		if r != nil && r.GetResource() != "" {
			q = q.Where(goqu.C("resource_id").Eq(r.GetResource()))
			q = q.Where(goqu.C("resource_type_id").Eq(r.GetResourceType()))
		}
	}

	if resourceReq, ok := req.(hasResourceListRequest); ok {
		r := resourceReq.GetResource()
		if r != nil {
			q = q.Where(goqu.C("resource_id").Eq(r.GetId().GetResource()))
			q = q.Where(goqu.C("resource_type_id").Eq(r.GetId().GetResourceType()))
		}
	}

	if entitlementReq, ok := req.(hasEntitlementListRequest); ok {
		e := entitlementReq.GetEntitlement()
		if e != nil {
			q = q.Where(goqu.C("entitlement_id").Eq(e.GetId()))
		}
	}

	if principalIdReq, ok := req.(hasPrincipalIdListRequest); ok {
		p := principalIdReq.GetPrincipalId()
		if p != nil {
			q = q.Where(goqu.C("principal_resource_id").Eq(p.GetResource()))
			q = q.Where(goqu.C("principal_resource_type_id").Eq(p.GetResourceType()))
		}
	}

	if principalResourceTypeIDsReq, ok := req.(hasPrincipalResourceTypeIDsListRequest); ok {
		p := principalResourceTypeIDsReq.GetPrincipalResourceTypeIds()
		if len(p) > 0 {
			q = q.Where(goqu.C("principal_resource_type_id").In(p))
		}
	}

	// If a sync is running, be sure we only select from the current values
	switch {
	case reqSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(reqSyncID))
	default:
		// Use cached sync run to avoid N+1 queries during pagination
		latestSyncRun, err := c.getCachedViewSyncRun(ctx)
		if err != nil {
			return nil, "", err
		}

		if latestSyncRun != nil {
			q = q.Where(goqu.C("sync_id").Eq(latestSyncRun.ID))
		}
	}

	// If a page token is provided, begin listing rows greater than or equal to the token
	if req.GetPageToken() != "" {
		q = q.Where(goqu.C("id").Gte(req.GetPageToken()))
	}

	// Clamp the page size
	pageSize := req.GetPageSize()
	if pageSize > maxPageSize || pageSize == 0 {
		pageSize = maxPageSize
	}
	// Always order by the row id
	q = q.Order(goqu.C("id").Asc())

	// Select 1 more than we asked for so we know if there is another page
	q = q.Limit(uint(pageSize + 1))

	query, args, err := q.ToSQL()
	if err != nil {
		return nil, "", err
	}

	// Start timing the query execution
	queryStartTime := time.Now()

	// Execute the query
	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	// Calculate the query duration
	queryDuration := time.Since(queryStartTime)

	// If the query took longer than the threshold, log a warning (rate-limited)
	if queryDuration > c.slowQueryThreshold {
		c.throttledWarnSlowQuery(ctx, query, queryDuration)
	}

	var unmarshalerOptions = proto.UnmarshalOptions{
		Merge:          true,
		DiscardUnknown: true,
	}
	var count uint32 = 0
	lastRow := 0
	var data sql.RawBytes
	var ret []T
	for rows.Next() {
		count++
		if count > pageSize {
			break
		}
		err := rows.Scan(&lastRow, &data)
		if err != nil {
			return nil, "", err
		}
		t := factory()
		err = unmarshalerOptions.Unmarshal(data, t)
		if err != nil {
			return nil, "", err
		}
		ret = append(ret, t)
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

// This is required for sync diffs to work.  Its not much slower.
var protoMarshaler = proto.MarshalOptions{Deterministic: true}

// prepareSingleConnectorObjectRow processes a single message and returns the prepared record.
func prepareSingleConnectorObjectRow[T proto.Message](
	c *C1File,
	msg T,
	extractFields func(m T) (goqu.Record, error),
) (*goqu.Record, error) {
	messageBlob, err := protoMarshaler.Marshal(msg)
	if err != nil {
		return nil, err
	}

	fields, err := extractFields(msg)
	if err != nil {
		return nil, err
	}
	if fields == nil {
		fields = goqu.Record{}
	}

	if _, idSet := fields["external_id"]; !idSet {
		idGetter, ok := any(msg).(protoHasID)
		if !ok {
			return nil, fmt.Errorf("unable to get ID for object")
		}
		fields["external_id"] = idGetter.GetId()
	}
	fields["data"] = messageBlob
	fields["sync_id"] = c.currentSyncID
	fields["discovered_at"] = time.Now().Format("2006-01-02 15:04:05.999999999")

	return &fields, nil
}

// prepareConnectorObjectRowsSerial prepares rows sequentially for bulk insertion.
func prepareConnectorObjectRowsSerial[T proto.Message](
	c *C1File,
	msgs []T,
	extractFields func(m T) (goqu.Record, error),
) ([]*goqu.Record, error) {
	rows := make([]*goqu.Record, len(msgs))
	for i, m := range msgs {
		row, err := prepareSingleConnectorObjectRow(c, m, extractFields)
		if err != nil {
			return nil, err
		}
		rows[i] = row
	}
	return rows, nil
}

// prepareConnectorObjectRowsParallel prepares rows for bulk insertion using parallel processing.
// For batches smaller than bulkPutParallelThreshold, it falls back to sequential processing.
func prepareConnectorObjectRowsParallel[T proto.Message](
	c *C1File,
	msgs []T,
	extractFields func(m T) (goqu.Record, error),
) ([]*goqu.Record, error) {
	if len(msgs) == 0 {
		return nil, nil
	}

	protoMarshallers := make([]proto.MarshalOptions, numWorkers)
	for i := range numWorkers {
		// Deterministic marshaling is required for sync diffs to work.  Its not much slower.
		protoMarshallers[i] = proto.MarshalOptions{Deterministic: true}
	}

	rows := make([]*goqu.Record, len(msgs))
	errs := make([]error, len(msgs))

	// Capture values that are the same for all rows (avoid repeated access)
	syncID := c.currentSyncID
	discoveredAt := time.Now().Format("2006-01-02 15:04:05.999999999")

	chunkSize := (len(msgs) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup

	for w := range numWorkers {
		start := w * chunkSize
		end := min(start+chunkSize, len(msgs))
		if start >= len(msgs) {
			break
		}

		wg.Add(1)
		go func(start, end int, worker int) {
			defer wg.Done()
			for i := start; i < end; i++ {
				m := msgs[i]

				messageBlob, err := protoMarshallers[worker].Marshal(m)
				if err != nil {
					errs[i] = err
					continue
				}

				fields, err := extractFields(m)
				if err != nil {
					errs[i] = err
					continue
				}
				if fields == nil {
					fields = goqu.Record{}
				}

				if _, idSet := fields["external_id"]; !idSet {
					idGetter, ok := any(m).(protoHasID)
					if !ok {
						errs[i] = fmt.Errorf("unable to get ID for object at index %d", i)
						continue
					}
					fields["external_id"] = idGetter.GetId()
				}
				fields["data"] = messageBlob
				fields["sync_id"] = syncID
				fields["discovered_at"] = discoveredAt
				rows[i] = &fields
			}
		}(start, end, w)
	}

	wg.Wait()

	// Check for errors (return first error encountered)
	for i, err := range errs {
		if err != nil {
			return nil, fmt.Errorf("error preparing row %d: %w", i, err)
		}
	}

	return rows, nil
}

// prepareConnectorObjectRows prepares the rows for bulk insertion.
// It uses parallel processing if the row count is greater than bulkPutParallelThreshold.
func prepareConnectorObjectRows[T proto.Message](
	c *C1File,
	msgs []T,
	extractFields func(m T) (goqu.Record, error),
) ([]*goqu.Record, error) {
	if len(msgs) > bulkPutParallelThreshold {
		return prepareConnectorObjectRowsParallel(c, msgs, extractFields)
	}
	return prepareConnectorObjectRowsSerial(c, msgs, extractFields)
}

// executeChunkedInsert executes the insert query in chunks.
func executeChunkedInsert(
	ctx context.Context,
	c *C1File,
	tableName string,
	rows []*goqu.Record,
	buildQueryFn func(*goqu.InsertDataset, []*goqu.Record) (*goqu.InsertDataset, error),
) error {
	chunkSize := insertChunkSize
	chunks := len(rows) / chunkSize
	if len(rows)%chunkSize != 0 {
		chunks++
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	var txError error

	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunkedRows := rows[start:end]

		// Create the base insert dataset
		insertDs := tx.Insert(tableName)

		// Apply the custom query building function
		insertDs, err = buildQueryFn(insertDs, chunkedRows)
		if err != nil {
			txError = err
			break
		}

		// Generate the SQL
		query, args, err := insertDs.ToSQL()
		if err != nil {
			txError = err
			break
		}

		// Execute the query
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			txError = err
			break
		}
	}

	if txError != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return errors.Join(rollbackErr, txError)
		}

		return fmt.Errorf("error executing chunked insert: %w", txError)
	}

	return tx.Commit()
}

func bulkPutConnectorObject[T proto.Message](
	ctx context.Context, c *C1File,
	tableName string,
	extractFields func(m T) (goqu.Record, error),
	msgs ...T,
) error {
	if len(msgs) == 0 {
		return nil
	}
	ctx, span := tracer.Start(ctx, "C1File.bulkPutConnectorObject")
	defer span.End()

	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	// Prepare rows
	rows, err := prepareConnectorObjectRows(c, msgs, extractFields)
	if err != nil {
		return err
	}

	// Define query building function
	buildQueryFn := func(insertDs *goqu.InsertDataset, chunkedRows []*goqu.Record) (*goqu.InsertDataset, error) {
		return insertDs.
			OnConflict(goqu.DoUpdate("external_id, sync_id", goqu.C("data").Set(goqu.I("EXCLUDED.data")))).
			Rows(chunkedRows).
			Prepared(true), nil
	}

	// Execute the insert
	return executeChunkedInsert(ctx, c, tableName, rows, buildQueryFn)
}

func bulkPutConnectorObjectIfNewer[T proto.Message](
	ctx context.Context, c *C1File,
	tableName string,
	extractFields func(m T) (goqu.Record, error),
	msgs ...T,
) error {
	if len(msgs) == 0 {
		return nil
	}
	ctx, span := tracer.Start(ctx, "C1File.bulkPutConnectorObjectIfNewer")
	defer span.End()

	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	// Prepare rows
	rows, err := prepareConnectorObjectRows(c, msgs, extractFields)
	if err != nil {
		return err
	}

	// Define query building function
	buildQueryFn := func(insertDs *goqu.InsertDataset, chunkedRows []*goqu.Record) (*goqu.InsertDataset, error) {
		return insertDs.
			OnConflict(goqu.DoUpdate("external_id, sync_id",
				goqu.Record{
					"data":          goqu.I("EXCLUDED.data"),
					"discovered_at": goqu.I("EXCLUDED.discovered_at"),
				}).Where(
				goqu.L("EXCLUDED.discovered_at > ?.discovered_at", goqu.I(tableName)),
			)).
			Rows(chunkedRows).
			Prepared(true), nil
	}

	// Execute the insert
	return executeChunkedInsert(ctx, c, tableName, rows, buildQueryFn)
}

func (c *C1File) getResourceObject(ctx context.Context, resourceID *v2.ResourceId, m *v2.Resource, syncID string) error {
	ctx, span := tracer.Start(ctx, "C1File.getResourceObject")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.From(resources.Name()).Prepared(true)
	q = q.Select("data")
	q = q.Where(goqu.C("resource_type_id").Eq(resourceID.GetResourceType()))
	q = q.Where(goqu.C("external_id").Eq(fmt.Sprintf("%s:%s", resourceID.GetResourceType(), resourceID.GetResource())))

	switch {
	case syncID != "":
		q = q.Where(goqu.C("sync_id").Eq(syncID))
	case c.currentSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	case c.viewSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(c.viewSyncID))
	default:
		var latestSyncRun *syncRun
		var err error
		latestSyncRun, err = c.getFinishedSync(ctx, 0, connectorstore.SyncTypeFull)
		if err != nil {
			return err
		}

		if latestSyncRun == nil {
			latestSyncRun, err = c.getLatestUnfinishedSync(ctx, connectorstore.SyncTypeAny)
			if err != nil {
				return err
			}
		}

		if latestSyncRun != nil {
			q = q.Where(goqu.C("sync_id").Eq(latestSyncRun.ID))
		}
	}

	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	data := make([]byte, 0)
	row := c.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&data)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(data, m)
	if err != nil {
		return err
	}

	return nil
}

func (c *C1File) getConnectorObject(ctx context.Context, tableName string, id string, syncID string, m proto.Message) error {
	ctx, span := tracer.Start(ctx, "C1File.getConnectorObject")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.From(tableName).Prepared(true)
	q = q.Select("data")
	q = q.Where(goqu.C("external_id").Eq(id))

	switch {
	case syncID != "":
		q = q.Where(goqu.C("sync_id").Eq(syncID))
	case c.currentSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	case c.viewSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(c.viewSyncID))
	default:
		var latestSyncRun *syncRun
		var err error
		latestSyncRun, err = c.getFinishedSync(ctx, 0, connectorstore.SyncTypeAny)
		if err != nil {
			return fmt.Errorf("error getting finished sync: %w", err)
		}

		if latestSyncRun == nil {
			latestSyncRun, err = c.getLatestUnfinishedSync(ctx, connectorstore.SyncTypeAny)
			if err != nil {
				return fmt.Errorf("error getting latest unfinished sync: %w", err)
			}
		}

		if latestSyncRun != nil {
			q = q.Where(goqu.C("sync_id").Eq(latestSyncRun.ID))
		}
	}

	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	data := make([]byte, 0)
	row := c.db.QueryRowContext(ctx, query, args...)
	err = row.Scan(&data)
	if err != nil {
		return err
	}

	err = proto.Unmarshal(data, m)
	if err != nil {
		return err
	}

	return nil
}
