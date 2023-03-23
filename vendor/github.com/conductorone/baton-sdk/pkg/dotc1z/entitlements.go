package dotc1z

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
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

func (c *C1File) ListEntitlements(ctx context.Context, request *v2.EntitlementsServiceListEntitlementsRequest) (*v2.EntitlementsServiceListEntitlementsResponse, error) {
	ctxzap.Extract(ctx).Debug("listing entitlements")
	objs, nextPageToken, err := c.listConnectorObjects(ctx, entitlements.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.Entitlement, 0, len(objs))
	for _, o := range objs {
		en := &v2.Entitlement{}
		err = proto.Unmarshal(o, en)
		if err != nil {
			return nil, err
		}
		ret = append(ret, en)
	}

	return &v2.EntitlementsServiceListEntitlementsResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) GetEntitlement(ctx context.Context, request *reader_v2.EntitlementsReaderServiceGetEntitlementRequest) (*v2.Entitlement, error) {
	ctxzap.Extract(ctx).Debug("fetching entitlement", zap.String("entitlement_id", request.EntitlementId))

	ret := &v2.Entitlement{}

	err := c.getConnectorObject(ctx, entitlements.Name(), request.EntitlementId, ret)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func (c *C1File) PutEntitlement(ctx context.Context, entitlement *v2.Entitlement) error {
	ctxzap.Extract(ctx).Debug("syncing entitlement", zap.String("entitlement_id", entitlement.Id))

	if entitlement.Resource == nil && entitlement.Resource.Id == nil {
		return fmt.Errorf("entitlements must have a non-nil resource")
	}

	query, args, err := c.putConnectorObjectQuery(ctx, entitlements.Name(), entitlement, goqu.Record{
		"resource_id":      entitlement.Resource.Id.Resource,
		"resource_type_id": entitlement.Resource.Id.ResourceType,
	})
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
