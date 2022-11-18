package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	ctx := context.Background()

	cliCmd := &cobra.Command{
		Use:     "baton",
		Short:   "baton is a utility for working with the output of a baton-based connector",
		Version: version,
	}

	cliCmd.PersistentFlags().StringP("file", "f", "sync.c1z", "The path to the c1z file to work with.")
	cliCmd.PersistentFlags().StringP("output-format", "o", "console", "The format to output results in: (console, json)")

	cliCmd.AddCommand(resourcesCmd())
	cliCmd.AddCommand(resourceTypesCmd())
	cliCmd.AddCommand(entitlementsCmd())
	cliCmd.AddCommand(grantsCmd())
	cliCmd.AddCommand(statsCmd())
	cliCmd.AddCommand(diffCmd())
	cliCmd.AddCommand(export())
	cliCmd.AddCommand(principalsCmd())

	err := cliCmd.ExecuteContext(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
