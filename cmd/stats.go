package main

import (
	"context"
	"sort"
	"strconv"

	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager"
	"github.com/conductorone/baton-sdk/pkg/logging"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func statsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Simple stats about the c1z",
		RunE:  runStats,
	}

	return cmd
}

func runStats(cmd *cobra.Command, args []string) error {
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

	counts, err := store.Stats(ctx)
	if err != nil {
		return err
	}

	statsTable := pterm.TableData{
		{"Type", "Count"},
	}

	var out [][]string
	for key, count := range counts {
		out = append(out, []string{
			key,
			strconv.Itoa(int(count)),
		})
	}

	sort.Slice(out, func(i int, j int) bool {
		return out[i][0] < out[j][0]
	})

	err = pterm.DefaultTable.WithHasHeader().WithData(append(statsTable, out...)).Render()
	if err != nil {
		return err
	}

	return nil
}
