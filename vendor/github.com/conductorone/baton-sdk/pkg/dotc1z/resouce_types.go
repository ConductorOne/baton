package dotc1z

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

const resourceTypesTableVersion = "1"
const resourceTypesTableName = "resource_types"
const resourcesTypesTableSchema = `
create table if not exists %s (
    id integer primary key,
    external_id text not null,
    data blob not null,
    sync_id text not null,
    discovered_at datetime not null
);
create unique index if not exists %s on %s (external_id, sync_id);`

var resourceTypes = (*resourceTypesTable)(nil)

type resourceTypesTable struct{}

func (r *resourceTypesTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), resourceTypesTableName)
}

func (r *resourceTypesTable) Version() string {
	return resourceTypesTableVersion
}

func (r *resourceTypesTable) Schema() (string, []interface{}) {
	return resourcesTypesTableSchema, []interface{}{
		r.Name(),
		fmt.Sprintf("idx_resource_types_external_sync_v%s", r.Version()),
		r.Name(),
	}
}

func (c *C1File) ListResourceTypes(ctx context.Context, request *v2.ResourceTypesServiceListResourceTypesRequest) (*v2.ResourceTypesServiceListResourceTypesResponse, error) {
	ctxzap.Extract(ctx).Debug("listing resource types")

	objs, nextPageToken, err := c.listConnectorObjects(ctx, resourceTypes.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.ResourceType, 0, len(objs))
	for _, o := range objs {
		rt := &v2.ResourceType{}
		err = proto.Unmarshal(o, rt)
		if err != nil {
			return nil, err
		}
		ret = append(ret, rt)
	}

	return &v2.ResourceTypesServiceListResourceTypesResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) GetResourceType(ctx context.Context, request *reader_v2.ResourceTypesReaderServiceGetResourceTypeRequest) (*v2.ResourceType, error) {
	ctxzap.Extract(ctx).Debug("fetching resource type", zap.String("resource_type_id", request.ResourceTypeId))

	ret := &v2.ResourceType{}

	err := c.getConnectorObject(ctx, resourceTypes.Name(), request.ResourceTypeId, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *C1File) PutResourceType(ctx context.Context, resourceType *v2.ResourceType) error {
	ctxzap.Extract(ctx).Debug("syncing resource type", zap.String("resource_type_id", resourceType.Id))

	query, args, err := c.putConnectorObjectQuery(ctx, resourceTypes.Name(), resourceType, nil)
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
