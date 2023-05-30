package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/conductorone/baton/pkg/storecache"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func resourcesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "List resources for the latest sync",
		RunE:  runResources,
	}

	addResourceTypeFlag(cmd)
	addSyncIDFlag(cmd)

	return cmd
}

func runResources(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), logging.WithLogFormat("console"), logging.WithLogLevel("error"))
	if err != nil {
		return err
	}
	c1zPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	resourceType, err := cmd.Flags().GetString(resourceTypeFlag)
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
