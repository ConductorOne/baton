package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

func grantsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grants",
		Short: "List grants",
		RunE:  runGrants,
	}

	addResourceTypeFlag(cmd)
	addResourceFlag(cmd)
	addEntitlementFlag(cmd)
	addSyncIDFlag(cmd)

	cmd.MarkFlagsMutuallyExclusive(resourceFlag, entitlementFlag)

	return cmd
}

func listGrantsForEntitlement(ctx context.Context, cmd *cobra.Command, store connectorstore.Reader, pageToken string) ([]*v2.Grant, string, error) {
	entitlementID, err := cmd.Flags().GetString(entitlementFlag)
	if err != nil {
		return nil, "", err
	}
	if entitlementID == "" {
		return nil, "", errors.New("--entitlement-id is required")
	}

	entitlement := &v2.Entitlement{Id: entitlementID}
	req := &reader_v2.GrantsReaderServiceListGrantsForEntitlementRequest{
		Entitlement: entitlement,
		PageToken:   pageToken,
	}
	resp, err := store.ListGrantsForEntitlement(ctx, req)
	if err != nil {
		return nil, "", err
	}

	return resp.List, resp.NextPageToken, nil
}

func listGrantsForResource(ctx context.Context, cmd *cobra.Command, store connectorstore.Reader, pageToken string) ([]*v2.Grant, string, error) {
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
	resp, err := store.ListGrants(ctx, req)
	if err != nil {
		return nil, "", err
	}

	return resp.List, resp.NextPageToken, nil
}

func listGrantsForResourceType(ctx context.Context, cmd *cobra.Command, store connectorstore.Reader, pageToken string) ([]*v2.Grant, string, error) {
	resourceTypeID, err := cmd.Flags().GetString(resourceTypeFlag)
	if err != nil {
		return nil, "", err
	}

	if resourceTypeID == "" {
		return nil, "", fmt.Errorf("--%s is required", resourceTypeFlag)
	}

	req := &reader_v2.GrantsReaderServiceListGrantsForResourceTypeRequest{
		ResourceTypeId: resourceTypeID,
		PageToken:      pageToken,
	}
	resp, err := store.ListGrantsForResourceType(ctx, req)
	if err != nil {
		return nil, "", err
	}

	return resp.List, resp.NextPageToken, nil
}

func listAllGrants(ctx context.Context, store connectorstore.Reader, pageToken string) ([]*v2.Grant, string, error) {
	req := &v2.GrantsServiceListGrantsRequest{
		PageToken: pageToken,
	}
	resp, err := store.ListGrants(ctx, req)
	if err != nil {
		return nil, "", err
	}

	return resp.List, resp.NextPageToken, nil
}

func runGrants(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), "console", "error")
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

	var grantOutputs []*v1.GrantOutput
	pageToken := ""
	for {
		var grants []*v2.Grant
		switch {
		case cmd.Flags().Changed(resourceFlag):
			grants, pageToken, err = listGrantsForResource(ctx, cmd, store, pageToken)
		case cmd.Flags().Changed(resourceTypeFlag):
			grants, pageToken, err = listGrantsForResourceType(ctx, cmd, store, pageToken)
		case cmd.Flags().Changed(entitlementFlag):
			grants, pageToken, err = listGrantsForEntitlement(ctx, cmd, store, pageToken)
		default:
			grants, pageToken, err = listAllGrants(ctx, store, pageToken)
		}
		if err != nil {
			return err
		}

		for _, g := range grants {
			en, err := sc.GetEntitlement(ctx, g.Entitlement.Id)
			if err != nil {
				return err
			}

			principal, err := sc.GetResource(ctx, g.Principal.Id)
			if err != nil {
				return err
			}

			resource, err := sc.GetResource(ctx, g.Entitlement.Resource.Id)
			if err != nil {
				return err
			}

			resourceType, err := sc.GetResourceType(ctx, g.Entitlement.Resource.Id.ResourceType)
			if err != nil {
				return err
			}

			grantOutputs = append(grantOutputs, &v1.GrantOutput{
				Grant:        g,
				Entitlement:  en,
				Principal:    principal,
				Resource:     resource,
				ResourceType: resourceType,
			})
		}

		if pageToken == "" {
			break
		}
	}

	err = outputManager.Output(ctx, &v1.GrantListOutput{Grants: grantOutputs})
	if err != nil {
		return err
	}

	return nil
}
