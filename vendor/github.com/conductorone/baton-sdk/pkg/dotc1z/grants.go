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

const grantsTableVersion = "1"
const grantsTableName = "grants"
const grantsTableSchema = `
create table if not exists %s (
    id integer primary key,
	resource_type_id text not null,
    resource_id text not null,
    entitlement_id text not null,
    principal_resource_type_id text not null,
    principal_resource_id text not null,
    external_id text not null,
    data blob not null,
    sync_id text not null,
    discovered_at datetime not null
);
create index if not exists %s on %s (resource_type_id, resource_id);
create index if not exists %s on %s (entitlement_id);
create index if not exists %s on %s (principal_resource_type_id, principal_resource_id);
create unique index if not exists %s on %s (external_id, sync_id);`

var grants = (*grantsTable)(nil)

type grantsTable struct{}

func (r *grantsTable) Version() string {
	return grantsTableVersion
}

func (r *grantsTable) Name() string {
	return fmt.Sprintf("v%s_%s", r.Version(), grantsTableName)
}

func (r *grantsTable) Schema() (string, []interface{}) {
	return grantsTableSchema, []interface{}{
		r.Name(),
		fmt.Sprintf("idx_resource_types_external_sync_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_entitlement_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_principal_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_external_sync_v%s", r.Version()),
		r.Name(),
	}
}

func (c *C1File) ListGrants(ctx context.Context, request *v2.GrantsServiceListGrantsRequest) (*v2.GrantsServiceListGrantsResponse, error) {
	ctxzap.Extract(ctx).Debug("listing grants")

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.Grant, 0, len(objs))
	for _, o := range objs {
		g := &v2.Grant{}
		err = proto.Unmarshal(o, g)
		if err != nil {
			return nil, err
		}
		ret = append(ret, g)
	}

	return &v2.GrantsServiceListGrantsResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) GetGrant(ctx context.Context, request *reader_v2.GrantsReaderServiceGetGrantRequest) (*reader_v2.GrantsReaderServiceGetGrantResponse, error) {
	ctxzap.Extract(ctx).Debug("fetching grant", zap.String("grant_id", request.GrantId))

	ret := &v2.Grant{}

	err := c.getConnectorObject(ctx, grants.Name(), request.GrantId, ret)
	if err != nil {
		return nil, err
	}

	return &reader_v2.GrantsReaderServiceGetGrantResponse{
		Grant: ret,
	}, nil
}

func (c *C1File) ListGrantsForEntitlement(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse, error) {
	ctxzap.Extract(ctx).Debug("listing grants for entitlement")

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.Grant, 0, len(objs))
	for _, o := range objs {
		en := &v2.Grant{}
		err = proto.Unmarshal(o, en)
		if err != nil {
			return nil, err
		}
		ret = append(ret, en)
	}

	return &reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) ListGrantsForPrincipal(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse, error) {
	ctxzap.Extract(ctx).Debug("listing grants for entitlement")

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.Grant, 0, len(objs))
	for _, o := range objs {
		en := &v2.Grant{}
		err = proto.Unmarshal(o, en)
		if err != nil {
			return nil, err
		}
		ret = append(ret, en)
	}

	return &reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) ListGrantsForResourceType(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForResourceTypeRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForResourceTypeResponse, error) {
	ctxzap.Extract(ctx).Debug("listing grants for resource type")

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, err
	}

	ret := make([]*v2.Grant, 0, len(objs))
	for _, o := range objs {
		en := &v2.Grant{}
		err = proto.Unmarshal(o, en)
		if err != nil {
			return nil, err
		}
		ret = append(ret, en)
	}

	return &reader_v2.GrantsReaderServiceListGrantsForResourceTypeResponse{
		List:          ret,
		NextPageToken: nextPageToken,
	}, nil
}

func (c *C1File) PutGrant(ctx context.Context, grant *v2.Grant) error {
	ctxzap.Extract(ctx).Debug("syncing grant", zap.String("grant_id", grant.Id))

	query, args, err := c.putConnectorObjectQuery(ctx, grants.Name(), grant, goqu.Record{
		"resource_type_id":           grant.Entitlement.Resource.Id.ResourceType,
		"resource_id":                grant.Entitlement.Resource.Id.Resource,
		"entitlement_id":             grant.Entitlement.Id,
		"principal_resource_type_id": grant.Principal.Id.ResourceType,
		"principal_resource_id":      grant.Principal.Id.Resource,
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
