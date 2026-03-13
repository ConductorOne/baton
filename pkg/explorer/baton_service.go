package explorer

import (
	"context"
	"fmt"
	"sync"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/storecache"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

// extractProfileFields extracts profile attributes from a resource's UserTrait annotations.
func extractProfileFields(r *v2.Resource) map[string]string {
	result := make(map[string]string)
	for _, ann := range r.GetAnnotations() {
		ut := &v2.UserTrait{}
		if err := ann.UnmarshalTo(ut); err != nil {
			continue
		}
		if ut.GetProfile() != nil {
			for k, v := range ut.GetProfile().GetFields() {
				result[k] = v.GetStringValue()
			}
		}
		for _, email := range ut.GetEmails() {
			if email.GetIsPrimary() && email.GetAddress() != "" {
				result["email"] = email.GetAddress()
				break
			}
		}
	}
	return result
}

type BatonService struct {
	syncID       string
	resourceType string
	store        *dotc1z.C1File
	storeCache   *storecache.StoreCache
	devMode      bool
	cache        sync.Map
}

// AccessCounts holds aggregated principal counts by resource type for a given resource.
type AccessCounts struct {
	TotalPrincipals int            `json:"total_principals"`
	CountsByType    map[string]int `json:"counts_by_type"`
}

func (b *BatonService) ensureSync(ctx context.Context) error {
	return b.store.ViewSync(ctx, b.syncID)
}

func (b *BatonService) getUserTraitTypes(ctx context.Context) (map[string]bool, error) {
	if v, ok := b.cache.Load("userTraitTypes"); ok {
		return v.(map[string]bool), nil
	}

	userTraitTypes := make(map[string]bool)
	pageToken := ""
	for {
		resp, err := b.store.ListResourceTypes(ctx, &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken})
		if err != nil {
			return nil, err
		}
		for _, rt := range resp.List {
			if len(rt.Traits) > 0 && rt.Traits[0] == v2.ResourceType_TRAIT_USER {
				userTraitTypes[rt.Id] = true
			}
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	b.cache.Store("userTraitTypes", userTraitTypes)
	return userTraitTypes, nil
}

func (b *BatonService) countResources(ctx context.Context, resourceTypeID string) (int, error) {
	cacheKey := "count:resources:" + resourceTypeID
	if v, ok := b.cache.Load(cacheKey); ok {
		return v.(int), nil
	}

	count := 0
	pageToken := ""
	for {
		resp, err := b.store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			ResourceTypeId: resourceTypeID,
			PageToken:      pageToken,
		})
		if err != nil {
			return 0, err
		}
		count += len(resp.List)
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	b.cache.Store(cacheKey, count)
	return count, nil
}

func (b *BatonService) GetEntitlements(ctx context.Context, pageToken string) (*v1.EntitlementListOutput, string, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	resp, err := b.store.ListEntitlements(ctx, &v2.EntitlementsServiceListEntitlementsRequest{PageToken: pageToken})
	if err != nil {
		return nil, "", err
	}

	var entitlements []*v1.EntitlementOutput
	for _, en := range resp.List {
		rt, err := b.storeCache.GetResourceType(ctx, en.Resource.Id.ResourceType)
		if err != nil {
			return nil, "", err
		}
		resource, err := b.storeCache.GetResource(ctx, en.Resource.Id)
		if err != nil {
			return nil, "", err
		}

		entitlements = append(entitlements, &v1.EntitlementOutput{
			Entitlement:  en,
			Resource:     resource,
			ResourceType: rt,
		})
	}

	return &v1.EntitlementListOutput{
		Entitlements: entitlements,
	}, resp.NextPageToken, nil
}

func (b *BatonService) GetResources(ctx context.Context, resourceTypeID string, pageToken string) (*v1.ResourceListOutput, string, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	filterType := resourceTypeID
	if filterType == "" {
		filterType = b.resourceType
	}

	resp, err := b.store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
		ResourceTypeId: filterType,
		PageToken:      pageToken,
	})
	if err != nil {
		return nil, "", err
	}

	var resources []*v1.ResourceOutput
	for _, r := range resp.List {
		rt, err := b.storeCache.GetResourceType(ctx, r.Id.ResourceType)
		if err != nil {
			return nil, "", err
		}
		var parent *v2.Resource

		if r.ParentResourceId != nil {
			parent, err = b.storeCache.GetResource(ctx, r.ParentResourceId)
			if err != nil {
				return nil, "", err
			}
		}

		resources = append(resources, &v1.ResourceOutput{
			Resource:     r,
			ResourceType: rt,
			Parent:       parent,
		})
	}

	return &v1.ResourceListOutput{
		Resources: resources,
	}, resp.NextPageToken, nil
}

func (b *BatonService) GetResourceTypes(ctx context.Context, pageToken string) (*v1.ResourceTypeListOutput, string, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	resp, err := b.store.ListResourceTypes(ctx, &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken})
	if err != nil {
		return nil, "", err
	}

	var resourceTypes []*v1.ResourceTypeOutput
	for _, rt := range resp.List {
		resourceTypes = append(resourceTypes, &v1.ResourceTypeOutput{ResourceType: rt})
	}

	return &v1.ResourceTypeListOutput{
		ResourceTypes: resourceTypes,
	}, resp.NextPageToken, nil
}

// GetAccess returns grants for a principal (user). Since users typically have <100 grants, this exhausts pagination.
func (b *BatonService) GetAccess(ctx context.Context, resourceType, resourceID string) (*v1.ResourceAccessListOutput, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, err
	}

	principal, err := b.storeCache.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	})
	if err != nil {
		return nil, err
	}

	var entitlements []*v2.Entitlement
	pageToken := ""
	for {
		req := reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest_builder{
			PrincipalId: &v2.ResourceId{
				ResourceType: resourceType,
				Resource:     resourceID,
			},
			PageToken: pageToken,
		}.Build()
		resp, err := b.store.ListGrantsForPrincipal(ctx, req)
		if err != nil {
			return nil, err
		}

		for _, g := range resp.List {
			en, err := b.storeCache.GetEntitlement(ctx, g.Entitlement.Id)
			if err != nil {
				return nil, err
			}
			entitlements = append(entitlements, en)
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
			resource, err := b.storeCache.GetResource(ctx, en.Resource.Id)
			if err != nil {
				return nil, err
			}

			rType, err := b.storeCache.GetResourceType(ctx, en.Resource.Id.ResourceType)
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

func (b *BatonService) GetResourceById(ctx context.Context, resourceType, resourceID string) (*ResourceDetailOutput, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, err
	}

	r, err := b.storeCache.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	})

	if err != nil {
		return nil, err
	}

	rt, err := b.storeCache.GetResourceType(ctx, resourceType)
	if err != nil {
		return nil, err
	}

	return &ResourceDetailOutput{
		Resource:     r,
		ResourceType: rt,
		Profile:      extractProfileFields(r),
	}, nil
}

// ResourceDetailOutput wraps a resource with its extracted profile fields for JSON serialization.
type ResourceDetailOutput struct {
	Resource     *v2.Resource      `json:"resource,omitempty"`
	ResourceType *v2.ResourceType  `json:"resource_type,omitempty"`
	Parent       *v2.Resource      `json:"parent,omitempty"`
	Profile      map[string]string `json:"profile,omitempty"`
}

func getResourceIdString(p *v2.Resource) string {
	return fmt.Sprintf("%s:%s", p.Id.ResourceType, p.Id.Resource)
}

type GrantsWithPrincipalResourceType struct {
	*v1.GrantOutput
	PrincipalResourceType *v2.ResourceType `json:"principal_resource_type,omitempty"`
}

type GrantsWithPrincipalResourceTypeListOutput struct {
	Grants []*GrantsWithPrincipalResourceType `json:"grants,omitempty"`
}

type ResourceAccessOutput struct {
	ResourceType *v2.ResourceType  `json:"resource_type,omitempty"`
	Resource     *v2.Resource      `json:"resource,omitempty"`
	Grants       []*v2.Grant       `json:"grants,omitempty"`
	Entitlements []*v2.Entitlement `json:"entitlements,omitempty"`
	Profile      map[string]string `json:"profile,omitempty"`
}

type ResourceAccessListOutput struct {
	Resource        *v2.Resource            `json:"resource,omitempty"`
	ResourceType    *v2.ResourceType        `json:"resource_type,omitempty"`
	PrincipalAccess []*ResourceAccessOutput `json:"access,omitempty"`
}

// GetAccessForResource returns grants for a resource (group/role), paginated by distinct principals.
// When computeCounts is true (first page), it performs a full scan to collect both paginated results
// and total counts by type in a single pass, avoiding a separate full scan for counts.
func (b *BatonService) GetAccessForResource(
	ctx context.Context, resourceType, resourceID string, pageSize int, pageToken string, computeCounts bool,
) (*ResourceAccessListOutput, string, *AccessCounts, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, "", nil, err
	}

	grantResource, err := b.storeCache.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	})
	if err != nil {
		return nil, "", nil, err
	}

	grantResourceType, err := b.storeCache.GetResourceType(ctx, resourceType)
	if err != nil {
		return nil, "", nil, err
	}

	if pageSize <= 0 {
		pageSize = 100
	}

	// Check cache for counts when computing them
	var cachedCounts *AccessCounts
	if computeCounts {
		cacheKey := fmt.Sprintf("accessCounts:%s:%s", resourceType, resourceID)
		if v, ok := b.cache.Load(cacheKey); ok {
			cachedCounts = v.(*AccessCounts)
			computeCounts = false // Already have counts, no need for full scan
		}
	}

	grantsByResource := make(map[string]*ResourceAccessOutput)
	pageFull := false
	nextPageToken := ""

	// For counts: track distinct principals by type across all pages
	var seenByType map[string]map[string]struct{}
	if computeCounts {
		seenByType = make(map[string]map[string]struct{})
	}

	currentToken := pageToken

	for {
		resp, err := b.store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			Resource: &v2.Resource{Id: &v2.ResourceId{
				ResourceType: resourceType,
				Resource:     resourceID,
			}},
			PageToken: currentToken,
		})
		if err != nil {
			return nil, "", nil, err
		}

		for _, g := range resp.List {
			rKey := getResourceIdString(g.Principal)

			// Always count for the full scan when computing counts
			if computeCounts {
				pType := g.Principal.Id.ResourceType
				if _, ok := seenByType[pType]; !ok {
					seenByType[pType] = make(map[string]struct{})
				}
				seenByType[pType][g.Principal.Id.Resource] = struct{}{}
			}

			// Only collect detailed access data while we haven't filled the page
			if !pageFull {
				var accessOutput *ResourceAccessOutput
				if rAccess, ok := grantsByResource[rKey]; ok {
					accessOutput = rAccess
				} else {
					if len(grantsByResource) >= pageSize {
						// We've hit pageSize distinct principals
						pageFull = true
						nextPageToken = resp.NextPageToken
						if !computeCounts {
							// Not computing counts, we can stop early
							break
						}
						continue
					}

					resource, err := b.storeCache.GetResource(ctx, g.Principal.Id)
					if err != nil {
						return nil, "", nil, err
					}

					rType, err := b.storeCache.GetResourceType(ctx, g.Principal.Id.ResourceType)
					if err != nil {
						return nil, "", nil, err
					}

					accessOutput = &ResourceAccessOutput{
						Resource:     resource,
						ResourceType: rType,
						Profile:      extractProfileFields(resource),
					}
				}

				en, err := b.storeCache.GetEntitlement(ctx, g.Entitlement.Id)
				if err != nil {
					return nil, "", nil, err
				}

				accessOutput.Grants = append(accessOutput.Grants, g)
				accessOutput.Entitlements = append(accessOutput.Entitlements, en)
				grantsByResource[rKey] = accessOutput
			}
		}

		// If not computing counts, stop once we have enough principals or no more pages
		if !computeCounts && (pageFull || resp.NextPageToken == "") {
			if !pageFull {
				nextPageToken = resp.NextPageToken
			}
			break
		}

		// If computing counts, we must exhaust all pages
		if resp.NextPageToken == "" {
			if !pageFull {
				nextPageToken = ""
			}
			break
		}

		currentToken = resp.NextPageToken
	}

	var outputs []*ResourceAccessOutput
	for _, o := range grantsByResource {
		outputs = append(outputs, o)
	}

	var counts *AccessCounts
	if computeCounts {
		counts = &AccessCounts{
			CountsByType: make(map[string]int),
		}
		for pType, ids := range seenByType {
			counts.CountsByType[pType] = len(ids)
			counts.TotalPrincipals += len(ids)
		}
		// Cache the counts for future requests
		cacheKey := fmt.Sprintf("accessCounts:%s:%s", resourceType, resourceID)
		b.cache.Store(cacheKey, counts)
	} else if cachedCounts != nil {
		counts = cachedCounts
	}

	return &ResourceAccessListOutput{
		Resource:        grantResource,
		ResourceType:    grantResourceType,
		PrincipalAccess: outputs,
	}, nextPageToken, counts, nil
}

func listPrincipalsForResource(
	ctx context.Context, resourceType, resourceID, pageToken string, sc *storecache.StoreCache,
) ([]*v2.Resource, string, error) {
	var ret []*v2.Resource

	resource := &v2.Resource{Id: &v2.ResourceId{
		ResourceType: resourceType,
		Resource:     resourceID,
	}}

	req := &v2.GrantsServiceListGrantsRequest{
		Resource:  resource,
		PageToken: pageToken,
	}

	resp, err := sc.Store().ListGrants(ctx, req)
	if err != nil {
		return nil, "", err
	}

	for _, g := range resp.List {
		p, err := sc.GetResource(ctx, g.Principal.Id)
		if err != nil {
			return nil, "", err
		}
		ret = append(ret, p)
	}

	return ret, resp.NextPageToken, nil
}

// GetPrincipals returns one page of user-trait principals for a resource.
func (b *BatonService) GetPrincipals(
	ctx context.Context, resourceType, resourceID string, pageToken string,
) (*v1.ResourceListOutput, string, error) {
	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	seenPrincipals := make(map[string]struct{})
	var outputs []*v1.ResourceOutput
	currentToken := pageToken

	// Fetch a single SDK page of grants and extract unique user principals.
	principals, nextToken, err := listPrincipalsForResource(ctx, resourceType, resourceID, currentToken, b.storeCache)
	if err != nil {
		return nil, "", err
	}

	for _, p := range principals {
		cacheKey := getResourceIdString(p)
		if _, ok := seenPrincipals[cacheKey]; !ok {
			rt, err := b.storeCache.GetResourceType(ctx, p.Id.ResourceType)
			if err != nil {
				return nil, "", err
			}

			var parent *v2.Resource
			if p.ParentResourceId != nil {
				parent, err = b.storeCache.GetResource(ctx, p.ParentResourceId)
				if err != nil {
					return nil, "", err
				}
			}

			if len(rt.Traits) > 0 && rt.Traits[0] == v2.ResourceType_TRAIT_USER {
				outputs = append(outputs, &v1.ResourceOutput{
					Resource:     p,
					ResourceType: rt,
					Parent:       parent,
				})
			}
			seenPrincipals[cacheKey] = struct{}{}
		}
	}

	return &v1.ResourceListOutput{
		Resources: outputs,
	}, nextToken, nil
}

type ResourceOutputWithCount struct {
	Resource     *v2.Resource     `json:"resource,omitempty"`
	ResourceType *v2.ResourceType `json:"resource_type,omitempty"`
	Parent       *v2.Resource     `json:"parent,omitempty"`
	UserCount    int              `json:"userCount"`
}

type ResourceListOutputWithCount struct {
	Resources []*ResourceOutputWithCount `json:"resources,omitempty"`
}

// GetResourcesWithPrincipalCount returns resources of a type with user principal counts. Results are cached.
func (b *BatonService) GetResourcesWithPrincipalCount(
	ctx context.Context, resourceType string,
) (*ResourceListOutputWithCount, error) {
	cacheKey := "principalCounts:" + resourceType
	if v, ok := b.cache.Load(cacheKey); ok {
		return v.(*ResourceListOutputWithCount), nil
	}

	if err := b.ensureSync(ctx); err != nil {
		return nil, err
	}

	userTraitTypes, err := b.getUserTraitTypes(ctx)
	if err != nil {
		return nil, err
	}

	var resources []*ResourceOutputWithCount
	pageToken := ""
	for {
		resp, err := b.store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			ResourceTypeId: resourceType,
			PageToken:      pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			rt, err := b.storeCache.GetResourceType(ctx, r.Id.ResourceType)
			if err != nil {
				return nil, err
			}
			var parent *v2.Resource

			if r.ParentResourceId != nil {
				parent, err = b.storeCache.GetResource(ctx, r.ParentResourceId)
				if err != nil {
					return nil, err
				}
			}

			seenPrincipals := make(map[string]struct{})
			grantPageToken := ""
			for {
				grantResp, err := b.store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
					Resource: &v2.Resource{Id: &v2.ResourceId{
						ResourceType: r.Id.ResourceType,
						Resource:     r.Id.Resource,
					}},
					PageToken: grantPageToken,
				})
				if err != nil {
					return nil, err
				}
				for _, g := range grantResp.List {
					if userTraitTypes[g.Principal.Id.ResourceType] {
						key := fmt.Sprintf("%s:%s", g.Principal.Id.ResourceType, g.Principal.Id.Resource)
						seenPrincipals[key] = struct{}{}
					}
				}
				if grantResp.NextPageToken == "" {
					break
				}
				grantPageToken = grantResp.NextPageToken
			}

			resources = append(resources, &ResourceOutputWithCount{
				Resource:     r,
				ResourceType: rt,
				Parent:       parent,
				UserCount:    len(seenPrincipals),
			})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	result := &ResourceListOutputWithCount{
		Resources: resources,
	}
	b.cache.Store(cacheKey, result)
	return result, nil
}
