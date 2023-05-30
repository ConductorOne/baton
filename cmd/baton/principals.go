package main

import (
	"context"
	"fmt"

	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func principalsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "principals",
		Short: "List principals",
		RunE:  runPrincipals,
	}

	addResourceFlag(cmd)
	addResourceTypeFlag(cmd)
	addEntitlementFlag(cmd)
	addSyncIDFlag(cmd)

	cmd.MarkFlagsRequiredTogether(resourceTypeFlag, resourceFlag)
	cmd.MarkFlagsMutuallyExclusive(resourceFlag, entitlementFlag)

	cmd.AddCommand(principalsCompareCmd())

	return cmd
}

func listPrincipalsForEntitlement(ctx context.Context, entitlementID string, sc *storecache.StoreCache, pageToken string) ([]*v2.Resource, string, error) {
	var ret []*v2.Resource

	entitlement := &v2.Entitlement{Id: entitlementID}
	req := &reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest{
		Entitlement: entitlement,
		PageToken:   pageToken,
	}
	resp, err := sc.Store().ListGrantsForEntitlement(ctx, req)
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

func listPrincipalsForResource(ctx context.Context, cmd *cobra.Command, sc *storecache.StoreCache, pageToken string) ([]*v2.Resource, string, error) {
	var ret []*v2.Resource

	resourceTypeID, err := cmd.Flags().GetString(resourceTypeFlag)
	if err != nil {
		return nil, "", err
	}
	resourceID, err := cmd.Flags().GetString(resourceFlag)
	if err != nil {
		return nil, "", err
	}
	if resourceTypeID == "" || resourceID == "" {
		return nil, "", fmt.Errorf("--%s and --%s are required", resourceTypeFlag, resourceFlag)
	}

	resource := &v2.Resource{Id: &v2.ResourceId{
		ResourceType: resourceTypeID,
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

func listAllPrincipals(ctx context.Context, sc *storecache.StoreCache, pageToken string) ([]*v2.Resource, string, error) {
	var ret []*v2.Resource

	req := &v2.GrantsServiceListGrantsRequest{
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

func getResourceIdString(p *v2.Resource) string {
	return fmt.Sprintf("%s:%s", p.Id.ResourceType, p.Id.Resource)
}

func runPrincipals(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), logging.WithLogFormat("console"), logging.WithLogLevel("error"))
	if err != nil {
		return err
	}
	c1zPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	outputFormat, err := cmd.Flags().GetString("output-format")
	if err != nil {
		return err
	}
	outputManager := output.NewManager(ctx, outputFormat)

	syncID, err := cmd.Flags().GetString("sync-id")
	if err != nil {
		return err
	}

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	if syncID != "" {
		err = store.ViewSync(ctx, syncID)
		if err != nil {
			return err
		}
	}

	sc := storecache.NewStoreCache(ctx, store)

	seenPrincipals := make(map[string]struct{})
	var outputs []*v1.ResourceOutput
	pageToken := ""
	for {
		var principals []*v2.Resource
		switch {
		case cmd.Flags().Changed(resourceFlag):
			principals, pageToken, err = listPrincipalsForResource(ctx, cmd, sc, pageToken)
		case cmd.Flags().Changed(entitlementFlag):
			var enID string
			enID, err = cmd.Flags().GetString(entitlementFlag)
			if err != nil {
				return err
			}

			principals, pageToken, err = listPrincipalsForEntitlement(ctx, enID, sc, pageToken)
		default:
			principals, pageToken, err = listAllPrincipals(ctx, sc, pageToken)
		}
		if err != nil {
			return err
		}

		for _, p := range principals {
			cacheKey := getResourceIdString(p)
			if _, ok := seenPrincipals[cacheKey]; !ok {
				resourceType, err := sc.GetResourceType(ctx, p.Id.ResourceType)
				if err != nil {
					return err
				}

				var parent *v2.Resource
				if p.ParentResourceId != nil {
					parent, err = sc.GetResource(ctx, p.ParentResourceId)
					if err != nil {
						return err
					}
				}

				outputs = append(outputs, &v1.ResourceOutput{
					Resource:     p,
					ResourceType: resourceType,
					Parent:       parent,
				})
				seenPrincipals[cacheKey] = struct{}{}
			}
		}

		if pageToken == "" {
			break
		}
	}

	err = outputManager.Output(ctx, &v1.ResourceListOutput{Resources: outputs})
	if err != nil {
		return err
	}

	return nil
}
