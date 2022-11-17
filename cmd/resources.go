package main

import (
	"context"

	v1 "github.com/conductorone/baton-cli/pb/baton_cli/v1"
	"github.com/conductorone/baton-cli/pkg/output"
	"github.com/conductorone/baton-cli/pkg/storecache"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func resourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "List resources for the latest sync",
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

	var resources []*v1.ResourceOutput
	pageToken := ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			ResourceTypeId: resourceType,
			PageToken:      pageToken,
		})
		if err != nil {
			return err
		}

		for _, r := range resp.List {
			rt, err := sc.GetResourceType(ctx, r.Id.ResourceType)
			if err != nil {
				return err
			}
			var parent *v2.Resource

			if r.ParentResourceId != nil {
				parent, err = sc.GetResource(ctx, r.ParentResourceId)
				if err != nil {
					return err
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

	err = outputManager.Output(ctx, &v1.ResourceListOutput{
		Resources: resources,
	})
	if err != nil {
		return err
	}

	return nil
}
