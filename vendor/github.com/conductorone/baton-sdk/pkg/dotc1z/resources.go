package dotc1z

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	"github.com/conductorone/baton-sdk/pkg/annotations"

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

func (r *resourcesTable) Migrations(ctx context.Context, db *goqu.Database) error {
	return nil
}

func (c *C1File) ListResources(ctx context.Context, request *v2.ResourcesServiceListResourcesRequest) (*v2.ResourcesServiceListResourcesResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListResources")
	defer span.End()

	objs, nextPageToken, err := c.listConnectorObjects(ctx, resources.Name(), request)
	if err != nil {
		return nil, fmt.Errorf("error listing resources: %w", err)
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
	ctx, span := tracer.Start(ctx, "C1File.GetResource")
	defer span.End()

	ret := &v2.Resource{}
	syncId, err := annotations.GetSyncIdFromAnnotations(request.GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error getting sync id from annotations for resource '%s': %w", request.ResourceId, err)
	}
	err = c.getResourceObject(ctx, request.ResourceId, ret, syncId)
	if err != nil {
		return nil, fmt.Errorf("error fetching resource '%s': %w", request.ResourceId, err)
	}

	return &reader_v2.ResourcesReaderServiceGetResourceResponse{
		Resource: ret,
	}, nil
}

func (c *C1File) PutResources(ctx context.Context, resourceObjs ...*v2.Resource) error {
	ctx, span := tracer.Start(ctx, "C1File.PutResources")
	defer span.End()

	return c.putResourcesInternal(ctx, bulkPutConnectorObject, resourceObjs...)
}

func (c *C1File) PutResourcesIfNewer(ctx context.Context, resourceObjs ...*v2.Resource) error {
	ctx, span := tracer.Start(ctx, "C1File.PutResourcesIfNewer")
	defer span.End()

	return c.putResourcesInternal(ctx, bulkPutConnectorObjectIfNewer, resourceObjs...)
}

type resourcePutFunc func(context.Context, *C1File, string, func(m *v2.Resource) (goqu.Record, error), ...*v2.Resource) error

func (c *C1File) putResourcesInternal(ctx context.Context, f resourcePutFunc, resourceObjs ...*v2.Resource) error {
	err := f(ctx, c, resources.Name(),
		func(resource *v2.Resource) (goqu.Record, error) {
			fields := goqu.Record{
				"resource_type_id": resource.Id.ResourceType,
				"external_id":      fmt.Sprintf("%s:%s", resource.Id.ResourceType, resource.Id.Resource),
			}

			// If we bulk insert some resources with parent ids and some without, goqu errors because of the different number of fields.
			if resource.ParentResourceId == nil {
				fields["parent_resource_type_id"] = nil
				fields["parent_resource_id"] = nil
			} else {
				fields["parent_resource_type_id"] = resource.ParentResourceId.ResourceType
				fields["parent_resource_id"] = resource.ParentResourceId.Resource
			}
			return fields, nil
		},
		resourceObjs...,
	)
	if err != nil {
		return err
	}
	c.dbUpdated = true
	return nil
}
