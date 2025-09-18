package dotc1z

import (
	"context"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"google.golang.org/protobuf/proto"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/annotations"
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
create index if not exists %s on %s (principal_resource_type_id, principal_resource_id);
create index if not exists %s on %s (entitlement_id, principal_resource_type_id, principal_resource_id);
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
		fmt.Sprintf("idx_grants_resource_type_id_resource_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_principal_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_entitlement_id_principal_id_v%s", r.Version()),
		r.Name(),
		fmt.Sprintf("idx_grants_external_sync_v%s", r.Version()),
		r.Name(),
	}
}

func (r *grantsTable) Migrations(ctx context.Context, db *goqu.Database) error {
	return nil
}

func (c *C1File) ListGrants(ctx context.Context, request *v2.GrantsServiceListGrantsRequest) (*v2.GrantsServiceListGrantsResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrants")
	defer span.End()

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants: %w", err)
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
	ctx, span := tracer.Start(ctx, "C1File.GetGrant")
	defer span.End()

	ret := &v2.Grant{}
	syncId, err := annotations.GetSyncIdFromAnnotations(request.GetAnnotations())
	if err != nil {
		return nil, fmt.Errorf("error getting sync id from annotations for grant '%s': %w", request.GrantId, err)
	}
	err = c.getConnectorObject(ctx, grants.Name(), request.GrantId, syncId, ret)
	if err != nil {
		return nil, fmt.Errorf("error fetching grant '%s': %w", request.GetGrantId(), err)
	}

	return &reader_v2.GrantsReaderServiceGetGrantResponse{
		Grant: ret,
	}, nil
}

func (c *C1File) ListGrantsForEntitlement(
	ctx context.Context,
	request *reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest,
) (*reader_v2.GrantsReaderServiceListGrantsForEntitlementResponse, error) {
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForEntitlement")
	defer span.End()

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants for entitlement '%s': %w", request.GetEntitlement().GetId(), err)
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
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForPrincipal")
	defer span.End()

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants for principal '%s': %w", request.GetPrincipalId(), err)
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
	ctx, span := tracer.Start(ctx, "C1File.ListGrantsForResourceType")
	defer span.End()

	objs, nextPageToken, err := c.listConnectorObjects(ctx, grants.Name(), request)
	if err != nil {
		return nil, fmt.Errorf("error listing grants for resource type '%s': %w", request.GetResourceTypeId(), err)
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

func (c *C1File) PutGrants(ctx context.Context, bulkGrants ...*v2.Grant) error {
	ctx, span := tracer.Start(ctx, "C1File.PutGrants")
	defer span.End()

	return c.putGrantsInternal(ctx, bulkPutConnectorObject, bulkGrants...)
}

func (c *C1File) PutGrantsIfNewer(ctx context.Context, bulkGrants ...*v2.Grant) error {
	ctx, span := tracer.Start(ctx, "C1File.PutGrantsIfNewer")
	defer span.End()

	return c.putGrantsInternal(ctx, bulkPutConnectorObjectIfNewer, bulkGrants...)
}

type grantPutFunc func(context.Context, *C1File, string, func(m *v2.Grant) (goqu.Record, error), ...*v2.Grant) error

func (c *C1File) putGrantsInternal(ctx context.Context, f grantPutFunc, bulkGrants ...*v2.Grant) error {
	err := f(ctx, c, grants.Name(),
		func(grant *v2.Grant) (goqu.Record, error) {
			return goqu.Record{
				"resource_type_id":           grant.Entitlement.Resource.Id.ResourceType,
				"resource_id":                grant.Entitlement.Resource.Id.Resource,
				"entitlement_id":             grant.Entitlement.Id,
				"principal_resource_type_id": grant.Principal.Id.ResourceType,
				"principal_resource_id":      grant.Principal.Id.Resource,
			}, nil
		},
		bulkGrants...,
	)
	if err != nil {
		return err
	}
	c.dbUpdated = true
	return nil
}

func (c *C1File) DeleteGrant(ctx context.Context, grantId string) error {
	ctx, span := tracer.Start(ctx, "C1File.DeleteGrant")
	defer span.End()

	err := c.validateSyncDb(ctx)
	if err != nil {
		return err
	}

	q := c.db.Delete(grants.Name())
	q = q.Where(goqu.C("external_id").Eq(grantId))
	if c.currentSyncID != "" {
		q = q.Where(goqu.C("sync_id").Eq(c.currentSyncID))
	}
	query, args, err := q.ToSQL()
	if err != nil {
		return err
	}

	_, err = c.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
