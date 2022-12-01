package main

import (
	"github.com/spf13/cobra"
)

type header int

func (h header) String() string {
	switch h {
	case headerType:
		return "Type"
	case headerLastName:
		return "Last Name"
	case headerFirstName:
		return "First Name"
	case headerUserID:
		return "User ID"
	case headerUserStatus:
		return "User Status"
	case headerEmailAddress:
		return "Email Address"
	case headerEntitlementDisplay:
		return "Entitlement Display Name"
	case headerEntitlement:
		return "Entitlement"
	case headerResourceType:
		return "Resource Type"
	case headerResourceName:
		return "Resource Name"
	case headerEntitlementDescription:
		return "Entitlement Description"
	case headerEntitlementSlug:
		return "Entitlement Slug"

	default:
		return "unknown"
	}
}

const (
	//nolint:deadcode,varcheck // Used for a starting point of iteration
	headerUnknown header = iota
	headerType
	headerLastName
	headerFirstName
	headerUserID
	headerUserStatus
	headerEmailAddress
	headerEntitlementDisplay
	headerEntitlement
	headerResourceType
	headerResourceName
	headerEntitlementDescription
	headerEntitlementSlug
	headerTerminator
)

func headers() []string {
	var ret []string
	for i := header(1); i < headerTerminator; i++ {
		ret = append(ret, i.String())
	}

	return ret
}

func export() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data from the C1Z for upload",
	}

	cmd.AddCommand(exportCSV())
	cmd.AddCommand(exportXLSX())

	return cmd
}
