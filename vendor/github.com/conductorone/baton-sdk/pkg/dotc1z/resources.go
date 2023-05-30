package dotc1z

import (
	"context"
	"fmt"

	c1zpb "github.com/conductorone/baton-sdk/pb/c1/c1z/v1"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

const resourcesTableVersion = "1"
const resourcesTableName = "resources"
const resourcesTableSchema = `
create table if not exists %s (
    id integer primary key,
    resource_type_id text not null,
    external_id text not null,
	parent_resource_type_id text,
	parent_resource_id text,
    data blob not null,
    sync_id text not null,
    discovered_at datetime not null
);
create index if not exists %s on %s (resource_type_id);
create index if not exists %s on %s (parent_resource_type_id, parent_resource_id);
create unique index if not exists %s on %s (external_id, sync_id);`

var resources = (*resourcesTable)(nil)

type resourcesTable struct{}

func (r *resourcesTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), resourcesTableName)
}

func (r *resourcesTable) Version() string {
	return resourcesTableVersion
}

func (r *resourcesTable) Schema() (string, []interface{}) {
	return resourcesTableSchema, []interface{}{
		r.Name(),
		fmt.Sprintf("idx_resources_resource_type_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_resources_external_sync_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_resources_parent_resource_id_v%s", r.Version()),
		r.Name(),
	}
}

func (c *C1File) ListResources(ctx context.Context, request *v2.ResourcesServiceListResourcesRequest) (*v2.ResourcesServiceListResourcesResponse, error) {
	ctxzap.Extract(ctx).Debug("listing resources")

	objs, nextPageToken, err := c.listConnectorObjects(ctx, resources.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.Resource, 0, len(objs))
	for _, o := range objs {
		rt := &v2.Resource{}
		err = proto.Unmarshal(o, rt)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}

	return &v2.ResourcesServiceListResourcesResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) GetResource(ctx context.Context, request *reader_v2.ResourcesReaderServiceGetResourceRequest) (*reader_v2.ResourcesReaderServiceGetResourceResponse, error) {
	ctxzap.Extract(ctx).Debug(
		"fetching resource",
		zap.String("resource_id", request.ResourceId.Resource),
		zap.String("resource_type_id", request.ResourceId.ResourceType),
	)

	ret := &v2.Resource{}
	annos := annotations.Annotations(request.GetAnnotations())
	syncDetails := &c1zpb.SyncDetails{}
	syncID := ""

	if ok, err := annos.Pick(syncDetails); err == nil && ok {
		syncID = syncDetails.GetId()
	}

	err := c.getResourceObject(ctx, request.ResourceId, ret, syncID)
	if err != nil {
		return nil, err
	}

	return &reader_v2.ResourcesReaderServiceGetResourceResponse{
		Resource: ret,
	}, nil
}

func (c *C1File) PutResource(ctx context.Context, resource *v2.Resource) error {
	ctxzap.Extract(ctx).Debug(
		"syncing resource",
		zap.String("resource_id", resource.Id.Resource),
		zap.String("resource_type_id", resource.Id.ResourceType),
	)

	updateRecord := goqu.Record{
		"resource_type_id": resource.Id.ResourceType,
		"external_id":      fmt.Sprintf("%s:%s", resource.Id.ResourceType, resource.Id.Resource),
	}

	if resource.ParentResourceId != nil {
		updateRecord["parent_resource_type_id"] = resource.ParentResourceId.ResourceType
		updateRecord["parent_resource_id"] = resource.ParentResourceId.Resource
	}

	query, args, err := c.putConnectorObjectQuery(ctx, resources.Name(), resource, updateRecord)
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
