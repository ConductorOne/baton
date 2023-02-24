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

func entitlementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entitlements",
		Short: "List entitlements",
		RunE:  runEntitlements,
	}

	addSyncIDFlag(cmd)

	return cmd
}

func runEntitlements(cmd *cobra.Command, args []string) error {
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

	var entitlements []*v1.EntitlementOutput
	pageToken := ""
	for {
		req := &v2.EntitlementsServiceListEntitlementsRequest{PageToken: pageToken}

		resp, err := store.ListEntitlements(ctx, req)
		if err != nil {
			return err
		}

		for _, en := range resp.List {
			rt, err := sc.GetResourceType(ctx, en.Resource.Id.ResourceType)
			if err != nil {
				return err
			}
			resource, err := sc.GetResource(ctx, en.Resource.Id)
			if err != nil {
				return err
			}

			entitlements = append(entitlements, &v1.EntitlementOutput{
				Entitlement:  en,
				Resource:     resource,
				ResourceType: rt,
			})
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	err = outputManager.Output(ctx, &v1.EntitlementListOutput{
		Entitlements: entitlements,
	})
	if err != nil {
		return err
	}

	return nil
}
