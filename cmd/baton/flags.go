package main

import (
	"github.com/spf13/cobra"
)

const (
	resourceTypeFlag = "resource-type"
	resourceFlag     = "resource"
	entitlementFlag  = "entitlement"
)

func addResourceTypeFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(resourceTypeFlag, "t", "", "The resource type to filter output by")
}

func addResourceFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(resourceFlag, "r", "", "The resource to filter output by")
}

func addEntitlementFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(entitlementFlag, "e", "", "The entitlement to filter output by")
}

func addSyncIDFlag(cmd *cobra.Command) {
	cmd.Flags().String("sync-id", "", "The sync ID to view data for. Will use the latest completed sync if not set.")
}
