package storecache

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type StoreCache struct {
	store         connectorstore.Reader
	resourceTypes sync.Map
	resources     sync.Map
	entitlements  sync.Map
	grants        sync.Map
}

func (f *StoreCache) getMissingResource(id *v2.ResourceId) *v2.Resource {
	return &v2.Resource{
		Id:          id,
		DisplayName: fmt.Sprintf("MISSING RESOURCE (%s:%s)", id.ResourceType, id.Resource),
	}
}

func (f *StoreCache) getMissingResourceType(id string) *v2.ResourceType {
	return &v2.ResourceType{
		Id:          id,
		DisplayName: fmt.Sprintf("MISSING RESOURCE TYPE (%s)", id),
	}
}

func (f *StoreCache) getMissingEntitlement(id string) *v2.Entitlement {
	return &v2.Entitlement{
		Id:          id,
		DisplayName: fmt.Sprintf("MISSING ENTITLEMENT (%s)", id),
	}
}

func (f *StoreCache) Store() connectorstore.Reader {
	return f.store
}

func (f *StoreCache) GetResourceType(ctx context.Context, id string) (*v2.ResourceType, error) {
	l := ctxzap.Extract(ctx)

	if id == "" {
		return nil, fmt.Errorf("resource type id must be set")
	}

	if v, ok := f.resourceTypes.Load(id); ok {
		return v.(*v2.ResourceType), nil
	}

	rt, err := f.store.GetResourceType(ctx, &reader_v2.ResourceTypesReaderServiceGetResourceTypeRequest{
		ResourceTypeId: id,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if rt == nil {
		l.Error("unable to find resource type", zap.String("resource_type_id", id))
		rt = f.getMissingResourceType(id)
	}

	f.resourceTypes.Store(id, rt)

	return rt, nil
}

func (f *StoreCache) getResourceKey(id *v2.ResourceId) string {
	return fmt.Sprintf("%s:%s", id.ResourceType, id.Resource)
}

func (f *StoreCache) GetResource(ctx context.Context, id *v2.ResourceId) (*v2.Resource, error) {
	l := ctxzap.Extract(ctx)

	cacheKey := f.getResourceKey(id)
	if id == nil {
		return nil, fmt.Errorf("resource id must be set")
	}

	if v, ok := f.resources.Load(cacheKey); ok {
		return v.(*v2.Resource), nil
	}

	resource, err := f.store.GetResource(ctx, &reader_v2.ResourceTypesReaderServiceGetResourceRequest{
		ResourceId: id,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if resource == nil {
		l.Error(
			"unable to find resource",
			zap.String("resource_type_id", id.ResourceType),
			zap.String("resource_id", id.Resource),
		)
		resource = f.getMissingResource(id)
	}

	f.resources.Store(cacheKey, resource)

	return resource, nil
}

func (f *StoreCache) GetEntitlement(ctx context.Context, id string) (*v2.Entitlement, error) {
	l := ctxzap.Extract(ctx)

	if id == "" {
		return nil, fmt.Errorf("entitlement id must be set")
	}

	if v, ok := f.entitlements.Load(id); ok {
		return v.(*v2.Entitlement), nil
	}

	entitlement, err := f.store.GetEntitlement(ctx, &reader_v2.EntitlementsReaderServiceGetEntitlementRequest{
		EntitlementId: id,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	if entitlement == nil {
		l.Error("unable to find entitlement", zap.String("entitlement_id", id))
		entitlement = f.getMissingEntitlement(id)
	}

	f.entitlements.Store(id, entitlement)

	return entitlement, nil
}

func (f *StoreCache) GetGrant(ctx context.Context, id string) (*v2.Grant, error) {
	if id == "" {
		return nil, fmt.Errorf("grant id must be set")
	}

	if v, ok := f.grants.Load(id); ok {
		return v.(*v2.Grant), nil
	}

	grant, err := f.store.GetGrant(ctx, &reader_v2.GrantsReaderServiceGetGrantRequest{
		GrantId: id,
	})
	if err != nil {
		return nil, err
	}

	f.grants.Store(id, grant)

	return grant, nil
}

func NewStoreCache(ctx context.Context, store connectorstore.Reader) *StoreCache {
	return &StoreCache{
		store: store,
	}
}
