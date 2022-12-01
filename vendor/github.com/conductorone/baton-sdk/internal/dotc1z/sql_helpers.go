package dotc1z

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

const maxPageSize = 100

type tableDescriptor interface {
	Name() string
	Schema() (string, []interface{})
	Version() string
}

type listRequest interface {
	proto.Message
	GetPageSize() uint32
	GetPageToken() string
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

type protoHasID interface {
	proto.Message
	GetId() string
}

// listConnectorObjects uses a connecter list request to fetch the corresponding data from the local db.
// It returns the raw bytes that need to be unmarshaled into the correct proto message.
func (c *C1File) listConnectorObjects(ctx context.Context, tableName string, req proto.Message) ([][]byte, string, error) {
	err := c.validateDb(ctx)
	if err != nil {
		return nil, "", err
	}

	// If this doesn't look like a list request, bail
	listReq, ok := req.(listRequest)
	if !ok {
		return nil, "", fmt.Errorf("c1file: invalid list request")
	}

	q := c.db.From(tableName).Prepared(true)
	q = q.Select("id", "data")

	// If the request allows filtering by resource type, apply the filter
	if resourceTypeReq, ok := req.(hasResourceTypeListRequest); ok {
		rt := resourceTypeReq.GetResourceTypeId()
		if rt != "" {
			q = q.Where(goqu.C("resource_type_id").Eq(rt))
		}
	} else if resourceIdReq, ok := req.(hasResourceIdListRequest); ok {
		r := resourceIdReq.GetResourceId()
		if r != nil && r.Resource != "" {
			q = q.Where(goqu.C("resource_id").Eq(r.Resource))
			q = q.Where(goqu.C("resource_type_id").Eq(r.ResourceType))
		}
	} else if resourceReq, ok := req.(hasResourceListRequest); ok {
		r := resourceReq.GetResource()
		if r != nil {
			q = q.Where(goqu.C("resource_id").Eq(r.Id.Resource))
			q = q.Where(goqu.C("resource_type_id").Eq(r.Id.ResourceType))
		}
	} else if entitlementReq, ok := req.(hasEntitlementListRequest); ok {
		e := entitlementReq.GetEntitlement()
		if e != nil {
			q = q.Where(goqu.C("entitlement_id").Eq(e.Id))
		}
	}

	// If a sync is running, be sure we only select from the current values
	switch {
	case c.currentSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	case c.viewSyncID != "":
		q = q.Where(goqu.C("sync_id").Eq(c.viewSyncID))
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
	pageSize := int(listReq.GetPageSize())
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

	rows, err := c.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	count := 0
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

func (c *C1File) putConnectorObjectQuery(ctx context.Context, tableName string, m proto.Message, fields goqu.Record) (string, []interface{}, error) {
	err := c.validateSyncDb(ctx)
	if err != nil {
		return "", nil, err
	}

	messageBlob, err := proto.Marshal(m)
	if err != nil {
		return "", nil, err
	}

	if fields == nil {
		fields = goqu.Record{}
	}

	if _, idSet := fields["external_id"]; !idSet {
		idGetter, ok := m.(protoHasID)
		if !ok {
			return "", nil, fmt.Errorf("unable to get ID for object")
		}
		fields["external_id"] = idGetter.GetId()
	}
	fields["data"] = messageBlob
	fields["sync_id"] = c.currentSyncID
	fields["discovered_at"] = time.Now().Format("2006-01-02 15:04:05.999999999")

	q := c.db.Insert(tableName).Prepared(true)
	q = q.Rows(fields)
	q = q.OnConflict(goqu.DoUpdate("external_id, sync_id", goqu.C("data").Set(goqu.I("EXCLUDED.data"))))

	return q.ToSQL()
}

func (c *C1File) getResourceObject(ctx context.Context, resourceID *v2.ResourceId, m *v2.Resource) error {
	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.From(resources.Name()).Prepared(true)
	q = q.Select("data")
	q = q.Where(goqu.C("resource_type_id").Eq(resourceID.ResourceType))
	q = q.Where(goqu.C("external_id").Eq(fmt.Sprintf("%s:%s", resourceID.ResourceType, resourceID.Resource)))

	switch {
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

func (c *C1File) getConnectorObject(ctx context.Context, tableName string, id string, m proto.Message) error {
	err := c.validateDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.From(tableName).Prepared(true)
	q = q.Select("data")
	q = q.Where(goqu.C("external_id").Eq(id))

	switch {
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
