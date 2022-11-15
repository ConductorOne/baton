package main

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

func main() {
	_ = context.Background()

	cliCmd := &cobra.Command{
		Use:     "baton",
		Short:   "baton is a utility for working with the output of a baton-based connector",
		Version: version,
	}

	cliCmd.PersistentFlags().StringP("file", "f", "sync.c1z", "The path to the c1z file to work with.")

	cliCmd.AddCommand(resourcesCmd())
	cliCmd.AddCommand(resourceTypesCmd())
	cliCmd.AddCommand(entitlementsCmd())
	cliCmd.AddCommand(grantsCmd())
	cliCmd.AddCommand(usersCmd())
	cliCmd.AddCommand(statsCmd())
	cliCmd.AddCommand(diffCmd())
	cliCmd.AddCommand(export())

	err := cliCmd.Execute()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
