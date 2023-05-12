package main

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/spf13/cobra"
)

func dumpDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "dump-db",
		Short:  "Dump the underlying database for the C1Z provided",
		RunE:   runDumpDB,
		Hidden: true,
	}

	cmd.Flags().String("out", "./sync.db", "The path to dump the database to")
	return cmd
}

func runDumpDB(cmd *cobra.Command, args []string) error {
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

	f, err := m.LoadRaw(ctx)
	if err != nil {
		return err
	}
	defer f.Close()

	dbFile, err := dotc1z.NewC1ZFileDecoder(f)
	if err != nil {
		return err
	}
	defer dbFile.Close()

	outPath, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}

	if outPath == "" {
		return fmt.Errorf("an output path is required")
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, dbFile)
	if err != nil {
		return err
	}

	return nil
}
