package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func entitlementsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entitlements",
		Short: "List entitlements",
		RunE:  runEntitlements,
	}

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

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}
	defer m.Close(ctx)

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	var entitlements []*v2.Entitlement
	pageToken := ""
	for {
		req := &v2.EntitlementsServiceListEntitlementsRequest{PageToken: pageToken}

		resp, err := store.ListEntitlements(ctx, req)
		if err != nil {
			return err
		}

		entitlements = append(entitlements, resp.List...)

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	resourcesTable := pterm.TableData{
		{"ID", "Resource", "Entitlement"},
	}
	for _, u := range entitlements {
		resourcesTable = append(resourcesTable, []string{
			u.Id,
			fmt.Sprintf("%s (%s)", u.Resource.DisplayName, u.Resource.Id.Resource),
			u.DisplayName,
		})
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
	if err != nil {
		return err
	}

	return nil
}
