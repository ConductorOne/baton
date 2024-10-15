package dotc1z

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	c1zpb "github.com/conductorone/baton-sdk/pb/c1/c1z/v1"
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

func (c *C1File) ListResources(ctx context.Context, request *v2.ResourcesServiceListResourcesRequest) (*v2.ResourcesServiceListResourcesResponse, error) {
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
	ret := &v2.Resource{}
	annos := annotations.Annotations(request.GetAnnotations())
	syncDetails := &c1zpb.SyncDetails{}
	syncID := ""

	if ok, err := annos.Pick(syncDetails); err == nil && ok {
		syncID = syncDetails.GetId()
	}

	err := c.getResourceObject(ctx, request.ResourceId, ret, syncID)
	if err != nil {
		return nil, fmt.Errorf("error fetching resource '%s': %w", request.ResourceId, err)
	}

	return &reader_v2.ResourcesReaderServiceGetResourceResponse{
		Resource: ret,
	}, nil
}

func (c *C1File) PutResources(ctx context.Context, resourceObjs ...*v2.Resource) error {
	err := c.db.WithTx(func(tx *goqu.TxDatabase) error {
		err := bulkPutConnectorObjectTx(ctx, c, tx, resources.Name(),
			func(resource *v2.Resource) (goqu.Record, error) {
				fields := goqu.Record{
					"resource_type_id": resource.Id.ResourceType,
					"external_id":      fmt.Sprintf("%s:%s", resource.Id.ResourceType, resource.Id.Resource),
				}

				if resource.ParentResourceId != nil {
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
		return nil
	})
	if err != nil {
		return err
	}
	c.dbUpdated = true
	return nil
}
