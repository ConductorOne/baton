package explorer

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/storecache"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

type BatonService struct {
	filePath     string
	syncID       string
	resourceType string
}

func (b BatonService) GetEntitlements(ctx context.Context) (*v1.EntitlementListOutput, error) {
	c1zPath := b.filePath

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}

	if b.syncID != "" {
		err = store.ViewSync(ctx, b.syncID)
		if err != nil {
			return nil, err
		}
	}

	sc := storecache.NewStoreCache(ctx, store)

	var entitlements []*v1.EntitlementOutput
	pageToken := ""
	for {
		req := &v2.EntitlementsServiceListEntitlementsRequest{PageToken: pageToken}

		resp, err := store.ListEntitlements(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, en := range resp.List {
			rt, err := sc.GetResourceType(ctx, en.Resource.Id.ResourceType)
			if err != nil {
				return nil, err
			}
			resource, err := sc.GetResource(ctx, en.Resource.Id)
			if err != nil {
				return nil, err
			}

			entitlements = append(entitlements, &v1.EntitlementOutput{
				Entitlement:  en,
				Resource:     resource,
				ResourceType: rt,
			})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return &v1.EntitlementListOutput{
		Entitlements: entitlements,
	}, err
}

func (b BatonService) GetResources(ctx context.Context) (*v1.ResourceListOutput, error) {
	m, err := manager.New(ctx, b.filePath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}

	err = store.ViewSync(ctx, "")
	if err != nil {
		return nil, err
	}

	sc := storecache.NewStoreCache(ctx, store)

	var resources []*v1.ResourceOutput
	pageToken := ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			ResourceTypeId: b.resourceType,
			PageToken:      pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			rt, err := sc.GetResourceType(ctx, r.Id.ResourceType)
			if err != nil {
				return nil, err
			}
			var parent *v2.Resource

			if r.ParentResourceId != nil {
				parent, err = sc.GetResource(ctx, r.ParentResourceId)
				if err != nil {
					return nil, err
				}
			}

			resources = append(resources, &v1.ResourceOutput{
				Resource:     r,
				ResourceType: rt,
				Parent:       parent,
			})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return &v1.ResourceListOutput{
		Resources: resources,
	}, nil
}

func (b BatonService) GetResourceTypes(ctx context.Context) (*v1.ResourceTypeListOutput, error) {
	m, err := manager.New(ctx, b.filePath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}

	if b.syncID != "" {
		err = store.ViewSync(ctx, b.syncID)
		if err != nil {
			return nil, err
		}
	}

	var resourceTypes []*v1.ResourceTypeOutput
	pageToken := ""
	for {
		resp, err := store.ListResourceTypes(ctx, &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken})
		if err != nil {
			return nil, err
		}

		for _, rt := range resp.List {
			resourceTypes = append(resourceTypes, &v1.ResourceTypeOutput{ResourceType: rt})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	return &v1.ResourceTypeListOutput{
		ResourceTypes: resourceTypes,
	}, nil
}

func (b BatonService) GetAccess(ctx context.Context, resourceType, resourceID string) (*v1.ResourceAccessListOutput, error) {
	m, err := manager.New(ctx, b.filePath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}

	sc := storecache.NewStoreCache(ctx, store)

	principal, err := sc.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	})
	if err != nil {
		return nil, err
	}

	var entitlements []*v2.Entitlement
	pageToken := ""
	for {
		resp, err := store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, g := range resp.List {
			if g.Principal.Id.ResourceType == resourceType && g.Principal.Id.Resource == resourceID {
				en, err := sc.GetEntitlement(ctx, g.Entitlement.Id)
				if err != nil {
					return nil, err
				}
				entitlements = append(entitlements, en)
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	entitlementsByResource := make(map[string]*v1.ResourceAccessOutput)

	for _, en := range entitlements {
		rKey := getResourceIdString(en.Resource)

		var accessOutput *v1.ResourceAccessOutput
		if rAccess, ok := entitlementsByResource[rKey]; ok {
			accessOutput = rAccess
		} else {
			resource, err := sc.GetResource(ctx, en.Resource.Id)
			if err != nil {
				return nil, err
			}

			rType, err := sc.GetResourceType(ctx, en.Resource.Id.ResourceType)
			if err != nil {
				return nil, err
			}

			accessOutput = &v1.ResourceAccessOutput{
				Resource:     resource,
				ResourceType: rType,
			}
		}

		accessOutput.Entitlements = append(accessOutput.Entitlements, en)
		entitlementsByResource[rKey] = accessOutput
	}

	var outputs []*v1.ResourceAccessOutput
	for _, o := range entitlementsByResource {
		outputs = append(outputs, o)
	}

	return &v1.ResourceAccessListOutput{
		Principal: principal,
		Access:    outputs,
	}, nil
}

func (b BatonService) GetResourceById(ctx context.Context, resourceType, resourceID string) (*v1.ResourceOutput, error) {
	m, err := manager.New(ctx, b.filePath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}

	err = store.ViewSync(ctx, "")
	if err != nil {
		return nil, err
	}

	sc := storecache.NewStoreCache(ctx, store)

	r, err := sc.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	})

	if err != nil {
		return nil, err
	}

	rt, err := sc.GetResourceType(ctx, resourceType)
	if err != nil {
		return nil, err
	}

	return &v1.ResourceOutput{
		Resource:     r,
		ResourceType: rt,
	}, nil
}

func listGrantsForResourceType(ctx context.Context, store connectorstore.Reader, pageToken string, resourceID string) ([]*v2.Grant, string, error) {
	req := &reader_v2.GrantsReaderServiceListGrantsForResourceTypeRequest{
		ResourceTypeId: resourceID,
		PageToken:      pageToken,
	}
	resp, err := store.ListGrantsForResourceType(ctx, req)
	if err != nil {
		return nil, "", err
	}

	return resp.List, resp.NextPageToken, nil
}

func getResourceIdString(p *v2.Resource) string {
	return fmt.Sprintf("%s:%s", p.Id.ResourceType, p.Id.Resource)
}

type GrantsWithPrincipalResourceType struct {
	*v1.GrantOutput
	PrincipalResourceType *v2.ResourceType `protobuf:"bytes,6,opt,name=principal_resource_type,json=principalResourceType,proto3" json:"principal_resource_type,omitempty"`
}

type GrantsWithPrincipalResourceTypeListOutput struct {
	Grants []*GrantsWithPrincipalResourceType `protobuf:"bytes,1,rep,name=grants,proto3" json:"grants,omitempty"`
}

type ResourceAccessOutput struct {
	ResourceType *v2.ResourceType  `protobuf:"bytes,1,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
	Resource     *v2.Resource      `protobuf:"bytes,2,opt,name=resource,proto3" json:"resource,omitempty"`
	Grants       []*v2.Grant       `protobuf:"bytes,3,rep,name=grants,proto3" json:"grants,omitempty"`
	Entitlements []*v2.Entitlement `protobuf:"bytes,4,rep,name=entitlements,proto3" json:"entitlements,omitempty"`
}

type ResourceAccessListOutput struct {
	Resource        *v2.Resource            `protobuf:"bytes,1,opt,name=resource,proto3" json:"resource,omitempty"`
	ResourceType    *v2.ResourceType        `protobuf:"bytes,2,opt,name=resource_type,json=resourceType,proto3" json:"resource_type,omitempty"`
	PrincipalAccess []*ResourceAccessOutput `protobuf:"bytes,3,rep,name=access,proto3" json:"access,omitempty"`
}

func (b BatonService) GetAccessForResource(ctx context.Context, resourceType, resourceID string) (*ResourceAccessListOutput, error) {
	m, err := manager.New(ctx, b.filePath)
	if err != nil {
		return nil, err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return nil, err
	}

	sc := storecache.NewStoreCache(ctx, store)

	grantResource, err := sc.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	})
	if err != nil {
		return nil, err
	}

	grantResourceType, err := sc.GetResourceType(ctx, resourceType)

	if err != nil {
		return nil, err
	}

	var resourceGrants []*v2.Grant
	pageToken := ""
	for {
		grants, nextToken, err := listGrantsForResourceType(ctx, store, pageToken, resourceType)
		if err != nil {
			return nil, err
		}

		for _, g := range grants {
			if g.Entitlement.Resource.Id.ResourceType == resourceType && g.Entitlement.Resource.Id.Resource == resourceID {
				resourceGrants = append(resourceGrants, g)
			}
		}

		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	grantsByResource := make(map[string]*ResourceAccessOutput)

	for _, g := range resourceGrants {
		rKey := getResourceIdString(g.Principal)

		var accessOutput *ResourceAccessOutput
		if rAccess, ok := grantsByResource[rKey]; ok {
			accessOutput = rAccess
		} else {
			resource, err := sc.GetResource(ctx, g.Principal.Id)
			if err != nil {
				return nil, err
			}

			rType, err := sc.GetResourceType(ctx, g.Principal.Id.ResourceType)
			if err != nil {
				return nil, err
			}

			accessOutput = &ResourceAccessOutput{
				Resource:     resource,
				ResourceType: rType,
			}
		}

		en, err := sc.GetEntitlement(ctx, g.Entitlement.Id)
		if err != nil {
			return nil, err
		}

		accessOutput.Grants = append(accessOutput.Grants, g)
		accessOutput.Entitlements = append(accessOutput.Entitlements, en)
		grantsByResource[rKey] = accessOutput
	}

	var outputs []*ResourceAccessOutput
	for _, o := range grantsByResource {
		outputs = append(outputs, o)
	}

	return &ResourceAccessListOutput{
		Resource:        grantResource,
		ResourceType:    grantResourceType,
		PrincipalAccess: outputs,
	}, nil
}