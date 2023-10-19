package main

import (
	"context"
	"fmt"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"
)

func optimizeDb() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "optimize",
		Short:  "Optimize the c1z file. This may result in a reduction in filesize.",
		RunE:   runOptimizeDb,
		Hidden: true,
	}

	return cmd
}

func runOptimizeDb(cmd *cobra.Command, args []string) error {
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

	err = store.Vacuum(ctx)
	if err != nil {
		return err
	}

	err = store.Close()
	if err != nil {
		return err
	}

	err = m.SaveC1Z(ctx)
	if err != nil {
		return err
	}

	fmt.Println("Optimized C1Z successfully.")
	return nil
}
