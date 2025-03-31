package dotc1z

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/conductorone/baton-sdk/pkg/annotations"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

const maxPageSize = 10000

var allTableDescriptors = []tableDescriptor{
	resourceTypes,
	resources,
	entitlements,
	grants,
	syncRuns,
	assets,
}

type tableDescriptor interface {
	Name() string
	Schema() (string, []interface{})
	Version() string
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
// It returns the raw bytes that need to be unmarshalled into the correct proto message.
func (c *C1File) listConnectorObjects(ctx context.Context, tableName string, req proto.Message) ([][]byte, string, error) {
	ctx, span := tracer.Start(ctx, "C1File.listConnectorObjects")
	defer span.End()

	err := c.validateDb(ctx)
	if err != nil {
		return nil, "", err
	}

	// If this doesn't look like a list request, bail
	listReq, ok := req.(listRequest)
	if !ok {
		return nil, "", fmt.Errorf("c1file: invalid list request")
	}

	annoSyncID, err := annotations.GetSyncIdFromAnnotations(listReq.GetAnnotations())
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
		if r != nil && r.Resource != "" {
			q = q.Where(goqu.C("resource_id").Eq(r.Resource))
			q = q.Where(goqu.C("resource_type_id").Eq(r.ResourceType))
		}
	}

	if resourceReq, ok := req.(hasResourceListRequest); ok {
		r := resourceReq.GetResource()
		if r != nil {
			q = q.Where(goqu.C("resource_id").Eq(r.Id.Resource))
			q = q.Where(goqu.C("resource_type_id").Eq(r.Id.ResourceType))
		}
	}

	if entitlementReq, ok := req.(hasEntitlementListRequest); ok {
		e := entitlementReq.GetEntitlement()
		if e != nil {
			q = q.Where(goqu.C("entitlement_id").Eq(e.Id))
		}
	}

	if principalIdReq, ok := req.(hasPrincipalIdListRequest); ok {
		p := principalIdReq.GetPrincipalId()
		if p != nil {
			q = q.Where(goqu.C("principal_resource_id").Eq(p.Resource))
			q = q.Where(goqu.C("principal_resource_type_id").Eq(p.ResourceType))
		}
	}

	// If a sync is running, be sure we only select from the current values
	switch {
	case reqSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(reqSyncID))
	default:
		var latestSyncRun *syncRun
		var err error
		latestSyncRun, err = c.getFinishedSync(ctx, 0)
		if err != nil {
			return nil, "", err
		}

		if latestSyncRun == nil {
			latestSyncRun, err = c.getLatestUnfinishedSync(ctx)
			if err != nil {
				return nil, "", err
			}
		}

		if latestSyncRun != nil {
			q = q.Where(goqu.C("sync_id").Eq(latestSyncRun.ID))
		}
	}

	// If a page token is provided, begin listing rows greater than or equal to the token
	if listReq.GetPageToken() != "" {
		q = q.Where(goqu.C("id").Gte(listReq.GetPageToken()))
	}

	// Clamp the page size
	pageSize := listReq.GetPageSize()
	if pageSize > maxPageSize || pageSize == 0 {
		pageSize = maxPageSize
	}
	// Always order by the row id
	q = q.Order(goqu.C("id").Asc())

	// Select 1 more than we asked for so we know if there is another page
	q = q.Limit(uint(pageSize + 1))

	var ret [][]byte

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

	var count uint32 = 0
	lastRow := 0
	for rows.Next() {
		count++
		if count > pageSize {
			break
		}
		rowId := 0
		data := make([]byte, 0)
		err := rows.Scan(&rowId, &data)
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

var protoMarshaler = proto.MarshalOptions{Deterministic: true}

func bulkPutConnectorObject[T proto.Message](ctx context.Context, c *C1File,
	tableName string,
	extractFields func(m T) (goqu.Record, error),
	msgs ...T) error {
	if len(msgs) == 0 {
		return nil
	}
	ctx, span := tracer.Start(ctx, "C1File.bulkPutConnectorObjectTx")
	defer span.End()

	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	rows := make([]*goqu.Record, len(msgs))
	for i, m := range msgs {
		messageBlob, err := protoMarshaler.Marshal(m)
		if err != nil {
			return err
		}

		fields, err := extractFields(m)
		if err != nil {
			return err
		}
		if fields == nil {
			fields = goqu.Record{}
		}

		if _, idSet := fields["external_id"]; !idSet {
			idGetter, ok := any(m).(protoHasID)
			if !ok {
				return fmt.Errorf("unable to get ID for object")
			}
			fields["external_id"] = idGetter.GetId()
		}
		fields["data"] = messageBlob
		fields["sync_id"] = c.currentSyncID
		fields["discovered_at"] = time.Now().Format("2006-01-02 15:04:05.999999999")
		rows[i] = &fields
	}
	chunkSize := 100
	chunks := len(rows) / chunkSize
	if len(rows)%chunkSize != 0 {
		chunks++
	}

	for i := 0; i < chunks; i++ {
		start := i * chunkSize
		end := (i + 1) * chunkSize
		if end > len(rows) {
			end = len(rows)
		}
		chunkedRows := rows[start:end]
		query, args, err := c.db.Insert(tableName).
			OnConflict(goqu.DoUpdate("external_id, sync_id", goqu.C("data").Set(goqu.I("EXCLUDED.data")))).
			Rows(chunkedRows).
			Prepared(true).
			ToSQL()
		if err != nil {
			return err
		}
		_, err = c.db.Exec(query, args...)
		if err != nil {
			return err
		}
	}

	return nil
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
	q = q.Where(goqu.C("resource_type_id").Eq(resourceID.ResourceType))
	q = q.Where(goqu.C("external_id").Eq(fmt.Sprintf("%s:%s", resourceID.ResourceType, resourceID.Resource)))

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
		latestSyncRun, err = c.getFinishedSync(ctx, 0)
		if err != nil {
			return err
		}

		if latestSyncRun == nil {
			latestSyncRun, err = c.getLatestUnfinishedSync(ctx)
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
		latestSyncRun, err = c.getFinishedSync(ctx, 0)
		if err != nil {
			return err
		}

		if latestSyncRun == nil {
			latestSyncRun, err = c.getLatestUnfinishedSync(ctx)
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
