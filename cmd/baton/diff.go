package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	v2 "github.com/conductorone/baton-sdk/pb/c1/connector/v2"
	"github.com/conductorone/baton-sdk/pkg/connectorstore"
	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func diffCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Perform a diff between sync runs",
		RunE:  runDiff,
	}

	return cmd
}

func runDiff(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), logging.WithLogFormat("console"), logging.WithLogLevel("error"))
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

	newSyncID, err := store.LatestSyncID(ctx, connectorstore.SyncTypeFull)
	if err != nil {
		return err
	}

	if newSyncID == "" {
		return fmt.Errorf("no syncs found - cannot diff")
	}

	oldSyncID, err := store.PreviousSyncID(ctx, connectorstore.SyncTypeFull)
	if err != nil {
		return err
	}

	if oldSyncID == "" {
		return fmt.Errorf("cannot diff single sync run")
	}

	rsDiff, err := bucketResources(ctx, store, oldSyncID, newSyncID)
	if err != nil {
		return err
	}

	enDiff, err := bucketEntitlements(ctx, store, oldSyncID, newSyncID)
	if err != nil {
		return err
	}

	grDiff, err := bucketGrants(ctx, store, oldSyncID, newSyncID)
	if err != nil {
		return err
	}

	out := &v1.C1ZDiffOutput{
		Resources:    rsDiff,
		Entitlements: enDiff,
		Grants:       grDiff,
	}

	diffBytes, err := protojson.Marshal(out)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, string(diffBytes))
	if err != nil {
		return err
	}

	return nil
}

func bucketResources(ctx context.Context, store *dotc1z.C1File, oldSyncID string, newSyncID string) (*v1.ResourceDiff, error) {
	ret := &v1.ResourceDiff{}

	oldResources := make(map[string]*v2.Resource)
	newResources := make(map[string]*v2.Resource)

	err := store.ViewSync(ctx, oldSyncID)
	if err != nil {
		return nil, err
	}

	pageToken := ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			oldResources[stringifyResourceID(r.Id)] = r
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	err = store.ViewSync(ctx, newSyncID)
	if err != nil {
		return nil, err
	}

	pageToken = ""
	for {
		resp, err := store.ListResources(ctx, &v2.ResourcesServiceListResourcesRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			newResources[stringifyResourceID(r.Id)] = r
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	for oldID, oldR := range oldResources {
		if newR, ok := newResources[oldID]; ok {
			equal, err := compareProto(ctx, oldR, newR)
			if err != nil {
				return nil, err
			}
			if !equal {
				ret.Modified = append(ret.Modified, newR)
			}
		} else {
			ret.Deleted = append(ret.Deleted, oldR)
		}
	}
	for newID, newR := range newResources {
		if _, ok := oldResources[newID]; !ok {
			ret.Created = append(ret.Created, newR)
		}
	}

	err = store.ViewSync(ctx, oldSyncID)
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func bucketEntitlements(ctx context.Context, store *dotc1z.C1File, oldSyncID string, newSyncID string) (*v1.EntitlementDiff, error) {
	ret := &v1.EntitlementDiff{}

	oldEntitlements := make(map[string]*v2.Entitlement)
	newEntitlements := make(map[string]*v2.Entitlement)

	err := store.ViewSync(ctx, oldSyncID)
	if err != nil {
		return nil, err
	}

	pageToken := ""
	for {
		resp, err := store.ListEntitlements(ctx, &v2.EntitlementsServiceListEntitlementsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			oldEntitlements[r.Id] = r
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	err = store.ViewSync(ctx, newSyncID)
	if err != nil {
		return nil, err
	}

	pageToken = ""
	for {
		resp, err := store.ListEntitlements(ctx, &v2.EntitlementsServiceListEntitlementsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			newEntitlements[r.Id] = r
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	for oldID, oldR := range oldEntitlements {
		if newR, ok := newEntitlements[oldID]; ok {
			equal, err := compareProto(ctx, oldR, newR)
			if err != nil {
				return nil, err
			}
			if !equal {
				ret.Modified = append(ret.Modified, newR)
			}
		} else {
			ret.Deleted = append(ret.Deleted, oldR)
		}
	}
	for newID, newR := range newEntitlements {
		if _, ok := oldEntitlements[newID]; !ok {
			ret.Created = append(ret.Created, newR)
		}
	}

	return ret, nil
}

func bucketGrants(ctx context.Context, store *dotc1z.C1File, oldSyncID string, newSyncID string) (*v1.GrantDiff, error) {
	ret := &v1.GrantDiff{}

	oldGrants := make(map[string]*v2.Grant)
	newGrants := make(map[string]*v2.Grant)

	err := store.ViewSync(ctx, oldSyncID)
	if err != nil {
		return nil, err
	}

	pageToken := ""
	for {
		resp, err := store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			oldGrants[r.Id] = r
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	err = store.ViewSync(ctx, newSyncID)
	if err != nil {
		return nil, err
	}

	pageToken = ""
	for {
		resp, err := store.ListGrants(ctx, &v2.GrantsServiceListGrantsRequest{
			PageToken: pageToken,
		})
		if err != nil {
			return nil, err
		}

		for _, r := range resp.List {
			newGrants[r.Id] = r
		}

		if resp.NextPageToken == "" {
			break
		}

		pageToken = resp.NextPageToken
	}

	for oldID, oldR := range oldGrants {
		if newR, ok := newGrants[oldID]; ok {
			equal, err := compareProto(ctx, oldR, newR)
			if err != nil {
				return nil, err
			}
			if !equal {
				ret.Modified = append(ret.Modified, newR)
			}
		} else {
			ret.Deleted = append(ret.Deleted, oldR)
		}
	}
	for newID, newR := range newGrants {
		if _, ok := oldGrants[newID]; !ok {
			ret.Created = append(ret.Created, newR)
		}
	}

	return ret, nil
}

func compareProto(ctx context.Context, oldR proto.Message, newR proto.Message) (bool, error) {
	oldBytes, err := protojson.Marshal(oldR)
	if err != nil {
		return false, err
	}
	var oldIface interface{}
	err = json.Unmarshal(oldBytes, &oldIface)
	if err != nil {
		return false, err
	}
	oldBytes, err = json.Marshal(oldIface)
	if err != nil {
		return false, err
	}

	newBytes, err := protojson.Marshal(newR)
	if err != nil {
		return false, err
	}
	var newIface interface{}
	err = json.Unmarshal(newBytes, &newIface)
	if err != nil {
		return false, err
	}
	newBytes, err = json.Marshal(newIface)
	if err != nil {
		return false, err
	}

	return bytes.Equal(oldBytes, newBytes), nil
}

func stringifyResourceID(rID *v2.ResourceId) string {
	return fmt.Sprintf("%s:%s", rID.ResourceType, rID.Resource)
}
