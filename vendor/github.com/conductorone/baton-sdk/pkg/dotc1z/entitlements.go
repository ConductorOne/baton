package dotc1z

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
)

const entitlementsTableVersion = "1"
const entitlementsTableName = "entitlements"
const entitlementsTableSchema = `
create table if not exists %s (
    id integer primary key,
    resource_type_id text not null,
    resource_id text not null,
    external_id text not null,
    data blob not null,
    sync_id text not null,
    discovered_at datetime not null
);
create index if not exists %s on %s (resource_type_id, resource_id);
create unique index if not exists %s on %s (external_id, sync_id);`

var entitlements = (*entitlementsTable)(nil)

type entitlementsTable struct{}

func (r *entitlementsTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), entitlementsTableName)
}

func (r *entitlementsTable) Version() string {
	return entitlementsTableVersion
}

func (r *entitlementsTable) Schema() (string, []interface{}) {
	return entitlementsTableSchema, []interface{}{
		r.Name(),
		fmt.Sprintf("idx_entitlements_resource_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_entitlements_external_sync_v%s", r.Version()),
		r.Name(),
	}
}

func (r *entitlementsTable) Migrations(ctx context.Context, db *goqu.Database) error {
	return nil
}

func (c *C1File) ListEntitlements(ctx context.Context, request *v2.EntitlementsServiceListEntitlementsRequest) (*v2.EntitlementsServiceListEntitlementsResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListEntitlements")
	defer span.End()

	objs, nextPageToken, err := listConnectorObjects(ctx, c, entitlements.Name(), request, func() *v2.Entitlement { return &v2.Entitlement{} })
	if err != nil {
		return nil, fmt.Errorf("error listing entitlements: %w", err)
	}

	return v2.EntitlementsServiceListEntitlementsResponse_builder{
		List:          objs,
		NextPageToken: nextPageToken,
	}.Build(), nil
}

func (c *C1File) GetEntitlement(ctx context.Context, request *reader_v2.EntitlementsReaderServiceGetEntitlementRequest) (*reader_v2.EntitlementsReaderServiceGetEntitlementResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.GetEntitlement")
	defer span.End()

	ret := &v2.Entitlement{}
	syncId, err := annotations.GetSyncIdFromAnnotations(request.GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error getting sync id from annotations for entitlement '%s': %w", request.GetEntitlementId(), err)
	}
	err = c.getConnectorObject(ctx, entitlements.Name(), request.GetEntitlementId(), syncId, ret)
	if err != nil {
		return nil, fmt.Errorf("error fetching entitlement '%s': %w", request.GetEntitlementId(), err)
	}

	return reader_v2.EntitlementsReaderServiceGetEntitlementResponse_builder{
		Entitlement: ret,
	}.Build(), nil
}

func (c *C1File) ListStaticEntitlements(ctx context.Context, request *v2.EntitlementsServiceListStaticEntitlementsRequest) (*v2.EntitlementsServiceListStaticEntitlementsResponse, error) {
	_, span := tracer.Start(ctx, "C1File.ListStaticEntitlements")
	defer span.End()

	return v2.EntitlementsServiceListStaticEntitlementsResponse_builder{
		List:          []*v2.Entitlement{},
		NextPageToken: "",
	}.Build(), nil
}

func (c *C1File) PutEntitlements(ctx context.Context, entitlementObjs ...*v2.Entitlement) error {
	ctx, span := tracer.Start(ctx, "C1File.PutEntitlements")
	defer span.End()

	return c.putEntitlementsInternal(ctx, bulkPutConnectorObject, entitlementObjs...)
}

func (c *C1File) PutEntitlementsIfNewer(ctx context.Context, entitlementObjs ...*v2.Entitlement) error {
	ctx, span := tracer.Start(ctx, "C1File.PutEntitlementsIfNewer")
	defer span.End()

	return c.putEntitlementsInternal(ctx, bulkPutConnectorObjectIfNewer, entitlementObjs...)
}

type entitlementPutFunc func(context.Context, *C1File, string, func(m *v2.Entitlement) (goqu.Record, error), ...*v2.Entitlement) error

func (c *C1File) putEntitlementsInternal(ctx context.Context, f entitlementPutFunc, entitlementObjs ...*v2.Entitlement) error {
	if c.readOnly {
		return ErrReadOnly
	}

	err := f(ctx, c, entitlements.Name(),
		func(entitlement *v2.Entitlement) (goqu.Record, error) {
			return goqu.Record{
				"resource_id":      entitlement.GetResource().GetId().GetResource(),
				"resource_type_id": entitlement.GetResource().GetId().GetResourceType(),
			}, nil
		},
		entitlementObjs...,
	)
	if err != nil {
		return err
	}
	c.dbUpdated = true
	return nil
}
