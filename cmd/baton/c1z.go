package main

import (
	"context"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"
)

func exportC1Z() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "c1z",
		Short: "Export latest generation to its own C1Z",
		RunE:  runExportC1Z,
	}

	cmd.Flags().String("out", "./latest.c1z", "The path to export the C1Z to")

	return cmd
}

func runExportC1Z(cmd *cobra.Command, args []string) error {
	ctx, err := logging.Init(context.Background(), "console", "error")
	if err != nil {
		return err
	}
	c1zPath, err := cmd.Flags().GetString("file")
	if err != nil {
		return err
	}

	outPath, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}

	m, err := manager.New(ctx, c1zPath)
	if err != nil {
		return err
	}

	store, err := m.LoadC1Z(ctx)
	if err != nil {
		return err
	}

	err = store.CloneSync(ctx, outPath, "")
	if err != nil {
		return err
	}

	return nil
}
