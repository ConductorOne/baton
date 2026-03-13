package explorer

import (
	"context"
	"fmt"
	"time"

	"github.com/google/cel-go/cel"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	v1 "github.com/conductorone/baton/pb/baton/v1"
)

const celSearchTimeout = 5 * time.Second

func newResourceCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("resource.display_name", cel.StringType),
		cel.Variable("resource.id", cel.StringType),
		cel.Variable("resource.type", cel.StringType),
		cel.Variable("resource.profile", cel.MapType(cel.StringType, cel.StringType)),
	)
}

func newGrantCELEnv() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("principal.display_name", cel.StringType),
		cel.Variable("principal.type", cel.StringType),
		cel.Variable("principal.id", cel.StringType),
		cel.Variable("entitlement.display_name", cel.StringType),
		cel.Variable("entitlement.slug", cel.StringType),
		cel.Variable("principal.profile", cel.MapType(cel.StringType, cel.StringType)),
	)
}

func compileCEL(env *cel.Env, expression string) (cel.Program, error) {
	ast, issues := env.Parse(expression)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("CEL parse error: %w", issues.Err())
	}

	checked, issues := env.Check(ast)
	if issues != nil && issues.Err() != nil {
		return nil, fmt.Errorf("CEL type-check error: %w", issues.Err())
	}

	prg, err := env.Program(checked)
	if err != nil {
		return nil, fmt.Errorf("CEL program error: %w", err)
	}

	return prg, nil
}

// SearchResources searches resources using a CEL expression. Iterates through all resources of the given type.
func (b *BatonService) SearchResources(
	ctx context.Context, expression string, resourceTypeID string, pageSize int, pageToken string,
) ([]*v1.ResourceOutput, string, error) {
	ctx, cancel := context.WithTimeout(ctx, celSearchTimeout)
	defer cancel()

	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	env, err := newResourceCELEnv()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create CEL environment: %w", err)
	}

	prg, err := compileCEL(env, expression)
	if err != nil {
		return nil, "", err
	}

	if pageSize <= 0 {
		pageSize = 100
	}

	var results []*v1.ResourceOutput
	currentToken := pageToken

	for {
		if ctx.Err() != nil {
			return nil, "", fmt.Errorf("search timed out")
		}

		resp, err := b.store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			ResourceTypeId: resourceTypeID,
			PageToken:      currentToken,
		})
		if err != nil {
			return nil, "", err
		}

		for _, r := range resp.List {
			if ctx.Err() != nil {
				return nil, "", fmt.Errorf("search timed out")
			}

			rt, err := b.storeCache.GetResourceType(ctx, r.Id.ResourceType)
			if err != nil {
				return nil, "", err
			}

			activation := map[string]any{
				"resource.display_name": r.DisplayName,
				"resource.id":           r.Id.Resource,
				"resource.type":         r.Id.ResourceType,
				"resource.profile":      extractProfileFields(r),
			}

			out, _, err := prg.Eval(activation)
			if err != nil {
				continue // skip items that error during eval
			}

			if out.Value() == true {
				var parent *v2.Resource
				if r.ParentResourceId != nil {
					parent, err = b.storeCache.GetResource(ctx, r.ParentResourceId)
					if err != nil {
						return nil, "", err
					}
				}

				results = append(results, &v1.ResourceOutput{
					Resource:     r,
					ResourceType: rt,
					Parent:       parent,
				})

				if len(results) >= pageSize {
					nextToken := resp.NextPageToken
					if nextToken == "" && len(resp.List) > 0 {
						// We consumed the whole page but hit the limit, no more continuation
						return results, "", nil
					}
					return results, nextToken, nil
				}
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		currentToken = resp.NextPageToken
	}

	return results, "", nil
}

// SearchGrants searches grants for a specific resource using a CEL expression.
func (b *BatonService) SearchGrants(
	ctx context.Context, expression string, resourceType, resourceID string, pageSize int, pageToken string,
) ([]*ResourceAccessOutput, string, error) {
	ctx, cancel := context.WithTimeout(ctx, celSearchTimeout)
	defer cancel()

	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	env, err := newGrantCELEnv()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create CEL environment: %w", err)
	}

	prg, err := compileCEL(env, expression)
	if err != nil {
		return nil, "", err
	}

	if pageSize <= 0 {
		pageSize = 100
	}

	matchedByResource := make(map[string]*ResourceAccessOutput)
	var matchOrder []string
	currentToken := pageToken

	for {
		if ctx.Err() != nil {
			return nil, "", fmt.Errorf("search timed out")
		}

		resp, err := b.store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			Resource: &v2.Resource{Id: &v2.ResourceId{
				ResourceType: resourceType,
				Resource:     resourceID,
			}},
			PageToken: currentToken,
		})
		if err != nil {
			return nil, "", err
		}

		for _, g := range resp.List {
			if ctx.Err() != nil {
				return nil, "", fmt.Errorf("search timed out")
			}

			principal, err := b.storeCache.GetResource(ctx, g.Principal.Id)
			if err != nil {
				return nil, "", err
			}

			en, err := b.storeCache.GetEntitlement(ctx, g.Entitlement.Id)
			if err != nil {
				return nil, "", err
			}

			activation := map[string]any{
				"principal.display_name":   principal.DisplayName,
				"principal.type":           g.Principal.Id.ResourceType,
				"principal.id":             g.Principal.Id.Resource,
				"entitlement.display_name": en.DisplayName,
				"entitlement.slug":         en.Slug,
				"principal.profile":        extractProfileFields(principal),
			}

			out, _, err := prg.Eval(activation)
			if err != nil {
				continue
			}

			if out.Value() == true {
				rKey := getResourceIdString(g.Principal)

				if existing, ok := matchedByResource[rKey]; ok {
					existing.Grants = append(existing.Grants, g)
					existing.Entitlements = append(existing.Entitlements, en)
				} else {
					principalRT, err := b.storeCache.GetResourceType(ctx, g.Principal.Id.ResourceType)
					if err != nil {
						return nil, "", err
					}

					accessOutput := &ResourceAccessOutput{
						Resource:     principal,
						ResourceType: principalRT,
						Grants:       []*v2.Grant{g},
						Entitlements: []*v2.Entitlement{en},
						Profile:      extractProfileFields(principal),
					}
					matchedByResource[rKey] = accessOutput
					matchOrder = append(matchOrder, rKey)

					if len(matchOrder) >= pageSize {
						nextToken := resp.NextPageToken
						return collectOrderedResults(matchedByResource, matchOrder), nextToken, nil
					}
				}
			}
		}

		// Also handle principal-based search using ListGrantsForPrincipal
		if resp.NextPageToken == "" {
			break
		}
		currentToken = resp.NextPageToken
	}

	return collectOrderedResults(matchedByResource, matchOrder), "", nil
}

// SearchGrantsForPrincipal searches grants where this resource is a principal using a CEL expression.
func (b *BatonService) SearchGrantsForPrincipal(
	ctx context.Context, expression string, resourceType, resourceID string, pageSize int, pageToken string,
) ([]*v1.ResourceAccessOutput, string, error) {
	ctx, cancel := context.WithTimeout(ctx, celSearchTimeout)
	defer cancel()

	if err := b.ensureSync(ctx); err != nil {
		return nil, "", err
	}

	env, err := newResourceCELEnv()
	if err != nil {
		return nil, "", fmt.Errorf("failed to create CEL environment: %w", err)
	}

	prg, err := compileCEL(env, expression)
	if err != nil {
		return nil, "", err
	}

	if pageSize <= 0 {
		pageSize = 100
	}

	entitlementsByResource := make(map[string]*v1.ResourceAccessOutput)
	var resultOrder []string
	currentToken := pageToken

	for {
		if ctx.Err() != nil {
			return nil, "", fmt.Errorf("search timed out")
		}

		req := reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest_builder{
			PrincipalId: &v2.ResourceId{
				ResourceType: resourceType,
				Resource:     resourceID,
			},
			PageToken: currentToken,
		}.Build()
		resp, err := b.store.ListGrantsForPrincipal(ctx, req)
		if err != nil {
			return nil, "", err
		}

		for _, g := range resp.List {
			en, err := b.storeCache.GetEntitlement(ctx, g.Entitlement.Id)
			if err != nil {
				return nil, "", err
			}

			resource, err := b.storeCache.GetResource(ctx, en.Resource.Id)
			if err != nil {
				return nil, "", err
			}

			rType, err := b.storeCache.GetResourceType(ctx, en.Resource.Id.ResourceType)
			if err != nil {
				return nil, "", err
			}

			activation := map[string]any{
				"resource.display_name": resource.DisplayName,
				"resource.id":           resource.Id.Resource,
				"resource.type":         resource.Id.ResourceType,
				"resource.profile":      extractProfileFields(resource),
			}

			out, _, err := prg.Eval(activation)
			if err != nil {
				continue
			}

			if out.Value() == true {
				rKey := getResourceIdString(en.Resource)

				if existing, ok := entitlementsByResource[rKey]; ok {
					existing.Entitlements = append(existing.Entitlements, en)
				} else {
					accessOutput := &v1.ResourceAccessOutput{
						Resource:     resource,
						ResourceType: rType,
						Entitlements: []*v2.Entitlement{en},
					}
					entitlementsByResource[rKey] = accessOutput
					resultOrder = append(resultOrder, rKey)

					if len(resultOrder) >= pageSize {
						var results []*v1.ResourceAccessOutput
						for _, k := range resultOrder {
							results = append(results, entitlementsByResource[k])
						}
						return results, resp.NextPageToken, nil
					}
				}
			}
		}

		if resp.NextPageToken == "" {
			break
		}
		currentToken = resp.NextPageToken
	}

	var results []*v1.ResourceAccessOutput
	for _, k := range resultOrder {
		results = append(results, entitlementsByResource[k])
	}
	return results, "", nil
}

func collectOrderedResults(m map[string]*ResourceAccessOutput, order []string) []*ResourceAccessOutput {
	results := make([]*ResourceAccessOutput, 0, len(order))
	for _, k := range order {
		results = append(results, m[k])
	}
	return results
}
