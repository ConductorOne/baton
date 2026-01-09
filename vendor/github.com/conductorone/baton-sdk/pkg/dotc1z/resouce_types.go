package dotc1z

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
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

func (r *resourceTypesTable) Migrations(ctx context.Context, db *goqu.Database) error {
	return nil
}

func (c *C1File) ListResourceTypes(ctx context.Context, request *v2.ResourceTypesServiceListResourceTypesRequest) (*v2.ResourceTypesServiceListResourceTypesResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListResourceTypes")
	defer span.End()

	ret, nextPageToken, err := listConnectorObjects(ctx, c, resourceTypes.Name(), request, func() *v2.ResourceType { return &v2.ResourceType{} })
	if err != nil {
		return nil, fmt.Errorf("error listing resource types: %w", err)
	}

	return v2.ResourceTypesServiceListResourceTypesResponse_builder{
		List:          ret,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

func (c *C1File) GetResourceType(ctx context.Context, request *reader_v2.ResourceTypesReaderServiceGetResourceTypeRequest) (*reader_v2.ResourceTypesReaderServiceGetResourceTypeResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.GetResourceType")
	defer span.End()

	ret := &v2.ResourceType{}
	syncId, err := annotations.GetSyncIdFromAnnotations(request.GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error getting sync id from annotations for resource type '%s': %w", request.GetResourceTypeId(), err)
	}
	err = c.getConnectorObject(ctx, resourceTypes.Name(), request.GetResourceTypeId(), syncId, ret)
	if err != nil {
		return nil, fmt.Errorf("error fetching resource type '%s': %w", request.GetResourceTypeId(), err)
	}

	return reader_v2.ResourceTypesReaderServiceGetResourceTypeResponse_builder{
		ResourceType: ret,
	}.Build(), nil
}

func (c *C1File) PutResourceTypes(ctx context.Context, resourceTypesObjs ...*v2.ResourceType) error {
	ctx, span := tracer.Start(ctx, "C1File.PutResourceTypes")
	defer span.End()

	return c.putResourceTypesInternal(ctx, bulkPutConnectorObject, resourceTypesObjs...)
}

func (c *C1File) PutResourceTypesIfNewer(ctx context.Context, resourceTypesObjs ...*v2.ResourceType) error {
	ctx, span := tracer.Start(ctx, "C1File.PutResourceTypesIfNewer")
	defer span.End()

	return c.putResourceTypesInternal(ctx, bulkPutConnectorObjectIfNewer, resourceTypesObjs...)
}

type resourceTypePutFunc func(context.Context, *C1File, string, func(m *v2.ResourceType) (goqu.Record, error), ...*v2.ResourceType) error

func (c *C1File) putResourceTypesInternal(ctx context.Context, f resourceTypePutFunc, resourceTypesObjs ...*v2.ResourceType) error {
	if c.readOnly {
		return ErrReadOnly
	}

	err := f(ctx, c, resourceTypes.Name(),
		func(resource *v2.ResourceType) (goqu.Record, error) {
			return nil, nil
		},
		resourceTypesObjs...,
	)
	if err != nil {
		return err
	}
	c.dbUpdated = true
	return nil
}
