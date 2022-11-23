package main

import (
	"context"
	"errors"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func accessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "List effective access for a user",
		RunE:  runAccess,
	}

	// Filter by resource
	cmd.Flags().String("resource-type-id", "", "Resource Type ID")
	cmd.Flags().String("resource-id", "", "Resource ID")
	cmd.MarkFlagsRequiredTogether("resource-type-id", "resource-id")

	return cmd
}

func runAccess(cmd *cobra.Command, args []string) error {
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

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	sc := storecache.NewStoreCache(ctx, store)

	resourceTypeID, err := cmd.Flags().GetString("resource-type-id")
	if err != nil {
		return err
	}
	resourceID, err := cmd.Flags().GetString("resource-id")
	if err != nil {
		return err
	}
	if resourceTypeID == "" || resourceID == "" {
		return errors.New("--resource-type-id and --resource-id are required")
	}

	principal, err := sc.GetResource(ctx, &v2.ResourceId{
		ResourceType: resourceTypeID,
		Resource:     resourceID,
	})
	if err != nil {
		return err
	}

	var entitlements []*v2.Entitlement
	pageToken := ""
	for {
		resp, err := store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return err
		}

		for _, g := range resp.List {
			if g.Principal.Id.ResourceType == resourceTypeID && g.Principal.Id.Resource == resourceID {
				en, err := sc.GetEntitlement(ctx, g.Entitlement.Id)
				if err != nil {
					return err
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
				return err
			}

			rType, err := sc.GetResourceType(ctx, en.Resource.Id.ResourceType)
			if err != nil {
				return err
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

	err = outputManager.Output(ctx, &v1.ResourceAccessListOutput{
		Principal: principal,
		Access:    outputs,
	})
	if err != nil {
		return err
	}

	return nil
}
