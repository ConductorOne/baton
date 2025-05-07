package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/conductorone/baton/pkg/output"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func syncsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "syncs",
		Short: "List the information for the various sync data stored in the C1Z",
		RunE:  runSyncList,
	}

	return cmd
}

func runSyncList(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), logging.WithLogFormat("console"), logging.WithLogLevel("error"))
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

	var syncRuns []*v1.SyncOutput
	pageToken := ""
	for {
		resp, nextPageToken, err := store.ListSyncRuns(ctx, pageToken, 100)
		if err != nil {
			return err
		}

		for _, sr := range resp {
			var startTime *timestamppb.Timestamp
			if sr.StartedAt != nil {
				startTime = timestamppb.New(*sr.StartedAt)
			}

			var endTime *timestamppb.Timestamp
			if sr.EndedAt != nil {
				endTime = timestamppb.New(*sr.EndedAt)
			}
			syncRuns = append(syncRuns, &v1.SyncOutput{
				Id:           sr.ID,
				StartedAt:    startTime,
				EndedAt:      endTime,
				SyncToken:    sr.SyncToken,
				SyncType:     string(sr.Type),
				ParentSyncId: sr.ParentSyncID,
			})
		}

		if nextPageToken == "" {
			break
		}

		pageToken = nextPageToken
	}

	err = outputManager.Output(ctx, &v1.SyncListOutput{
		Syncs: syncRuns,
	})
	if err != nil {
		return err
	}

	return nil
}
