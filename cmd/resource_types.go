package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func resourceTypesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resource-types",
		Short: "List resource types for the latest (or current) sync",
		RunE:  runResourceTypes,
	}

	return cmd
}

func runResourceTypes(cmd *cobra.Command, args []string) error {
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

	var resourceTypes []*v1.ResourceTypeOutput
	pageToken := ""
	for {
		resp, err := store.ListResourceTypes(ctx, &v2.ResourceTypesServiceListResourceTypesRequest{PageToken: pageToken})
		if err != nil {
			return err
		}

		for _, rt := range resp.List {
			resourceTypes = append(resourceTypes, &v1.ResourceTypeOutput{ResourceType: rt})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	err = outputManager.Output(ctx, &v1.ResourceTypeListOutput{
		ResourceTypes: resourceTypes,
	})
	if err != nil {
		return err
	}

	return nil
}
