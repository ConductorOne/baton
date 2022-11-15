package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/pterm/pterm"
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

	// Filter by resource
	cmd.Flags().String("resource-type-id", "", "Resource Type ID")
	cmd.Flags().String("resource-id", "", "Resource ID")
	cmd.MarkFlagsRequiredTogether("resource-type-id", "resource-id")

	// Filter by entitlement
	cmd.Flags().String("entitlement-id", "", "Entitlement ID")

	cmd.MarkFlagsMutuallyExclusive("resource-id", "entitlement-id")

	return cmd
}

func listGrantsForEntitlement(ctx context.Context, cmd *cobra.Command, store connectorstore.Reader, pageToken string) ([]*v2.Grant, string, error) {
	entitlementID, err := cmd.Flags().GetString("entitlement-id")
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
	resourceTypeID, err := cmd.Flags().GetString("resource-type-id")
	if err != nil {
		return nil, "", err
	}
	resourceID, err := cmd.Flags().GetString("resource-id")
	if err != nil {
		return nil, "", err
	}
	if resourceTypeID == "" || resourceID == "" {
		return nil, "", errors.New("--resource-type-id and --resource-id are required")
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

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	var grants []*v2.Grant
	pageToken := ""
	for {
		switch {
		case cmd.Flags().Changed("resource-id"):
			grants, pageToken, err = listGrantsForResource(ctx, cmd, store, pageToken)
		case cmd.Flags().Changed("entitlement-id"):
			grants, pageToken, err = listGrantsForEntitlement(ctx, cmd, store, pageToken)
		default:
			grants, pageToken, err = listAllGrants(ctx, store, pageToken)
		}
		if err != nil {
			return err
		}

		resourcesTable := pterm.TableData{
			{"Resource", "Entitlement", "Principal"},
		}
		for _, u := range grants {
			en, err := store.GetEntitlement(ctx, &reader_v2.EntitlementsReaderServiceGetEntitlementRequest{
				EntitlementId: u.Entitlement.Id,
			})
			if err != nil {
				return err
			}

			principal, err := store.GetResource(ctx, &reader_v2.ResourceTypesReaderServiceGetResourceRequest{
				ResourceId: u.Principal.Id,
			})
			if err != nil {
				return err
			}

			resourcesTable = append(resourcesTable, []string{
				fmt.Sprintf("%s: %s", en.Resource.Id.ResourceType, en.Resource.DisplayName),
				en.DisplayName,
				principal.DisplayName,
			})
		}

		err = pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
		if err != nil {
			return err
		}

		if pageToken == "" {
			break
		}

		result, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Next Page?").WithMultiLine(false).Show()
		if result == "n" {
			break
		}
	}

	return nil
}
