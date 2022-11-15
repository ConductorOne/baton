package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	reader_v2 "github.com/conductorone/baton-sdk/pb/c1/reader/v2"
)

func resourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "List resources for the latest (or current) sync",
		RunE:  runResources,
	}

	cmd.Flags().String("resource-type", "", "The resource type to list resources for")

	return cmd
}

func runResources(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), "console", "error")
	if err != nil {
		return err
	}
	c1zPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	resourceType, err := cmd.Flags().GetString("resource-type")
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

	var resources []*v2.Resource
	pageToken := ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{ResourceTypeId: resourceType, PageToken: pageToken})
		if err != nil {
			return err
		}

		resources = append(resources, resp.List...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	resourcesTable := pterm.TableData{
		{"ID", "Display Name", "Resource Type", "Parent Resource"},
	}
	for _, u := range resources {
		rType, err := store.GetResourceType(ctx, &reader_v2.ResourceTypesReaderServiceGetResourceTypeRequest{
			ResourceTypeId: u.Id.ResourceType,
		})
		if err != nil {
			return err
		}

		parentResourceText := "-"
		if u.ParentResourceId != nil {
			parentResource, err := store.GetResource(ctx, &reader_v2.ResourceTypesReaderServiceGetResourceRequest{
				ResourceId: u.ParentResourceId,
			})
			if err != nil {
				return err
			}
			parentResourceText = fmt.Sprintf(
				"%s (%s - %s)",
				parentResource.DisplayName,
				parentResource.Id.ResourceType,
				parentResource.Id.Resource,
			)
		}

		resourcesTable = append(resourcesTable, []string{
			u.Id.Resource,
			u.DisplayName,
			rType.DisplayName,
			parentResourceText,
		})
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
	if err != nil {
		return err
	}

	return nil
}
