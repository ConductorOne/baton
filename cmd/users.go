package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
)

func usersCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "List user resources with more detail",
		RunE:  runUsers,
	}

	return cmd
}

func runUsers(cmd *cobra.Command, args []string) error {
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

	var resources []*v2.Resource
	pageToken := ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			ResourceTypeId: "user",
			PageToken:      pageToken,
		})
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
		{"ID", "Display Name", "Login", "Email", "Profile Image"},
	}
	for _, u := range resources {
		userTrait := &v2.UserTrait{}
		annos := annotations.Annotations(u.Annotations)
		ok, err := annos.Pick(userTrait)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}

		var primaryEmail string
		for _, e := range userTrait.Emails {
			if e.IsPrimary {
				primaryEmail = e.Address
				break
			}
		}

		var iconFname string
		if userTrait.Icon != nil {
			iconContentType, iconReader, err := store.GetAsset(ctx, &v2.AssetServiceGetAssetRequest{Asset: &v2.AssetRef{Id: userTrait.Icon.Id}})
			if err != nil {
				return err
			}

			var fnamePattern string
			switch iconContentType {
			case "image/png":
				fnamePattern = "*.png"
			case "image/jpeg":
				fnamePattern = "*.jpeg"
			default:
				return fmt.Errorf("unexpected content type %s", iconContentType)
			}
			f, err := os.CreateTemp("", fnamePattern)
			if err != nil {
				return err
			}
			iconFname = f.Name()
			_, err = io.Copy(f, iconReader)
			if err != nil {
				return err
			}
			_ = f.Close()
		}

		var login string
		if userTrait.Profile != nil {
			login = userTrait.Profile.Fields["login"].GetStringValue()
		}

		resourcesTable = append(resourcesTable, []string{
			u.Id.Resource,
			u.DisplayName,
			login,
			primaryEmail,
			iconFname,
		})
	}

	err = pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
	if err != nil {
		return err
	}

	return nil
}
