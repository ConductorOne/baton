package output

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	v1 "github.com/conductorone/baton/pb/baton/v1"
	"github.com/pterm/pterm"
	"github.com/pterm/pterm/putils"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type consoleManager struct{}

func (c *consoleManager) Output(ctx context.Context, out interface{}) error {
	switch obj := out.(type) {
	case *v1.ResourceTypeListOutput:
		return c.outputResourceTypes(obj)

	case *v1.ResourceListOutput:
		return c.outputResources(obj)

	case *v1.EntitlementListOutput:
		return c.outputEntitlements(obj)

	case *v1.GrantListOutput:
		return c.outputGrants(obj)

	case *v1.ResourceAccessListOutput:
		return c.outputResourceAccess(obj)

	case *v1.PrincipalsCompareOutput:
		return c.outputPrincipalsCompare(obj)

	case *v1.SyncListOutput:
		return c.outputSyncRuns(obj)

	default:
		return fmt.Errorf("unexpected output model")
	}
}

func (c *consoleManager) formatTimestamp(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}

	return ts.AsTime().Format(time.RFC3339)
}

func (c *consoleManager) outputSyncRuns(out *v1.SyncListOutput) error {
	syncsTable := pterm.TableData{
		{"ID", "Started At", "Ended At", "Type", "Parent ID", "Token"},
	}

	for _, o := range out.Syncs {
		syncsTable = append(syncsTable, []string{
			o.Id,
			c.formatTimestamp(o.StartedAt),
			c.formatTimestamp(o.EndedAt),
			o.SyncType,
			o.ParentSyncId,
			o.SyncToken,
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithData(syncsTable).Render()
	if err != nil {
		return err
	}

	return nil
}

func (c *consoleManager) outputResourceTypes(out *v1.ResourceTypeListOutput) error {
	resourceTypesTable := pterm.TableData{
		{"ID", "Display Name", "Traits"},
	}

	for _, o := range out.ResourceTypes {
		var traits []string
		for _, t := range o.ResourceType.Traits {
			traits = append(traits, t.String())
		}

		resourceTypesTable = append(resourceTypesTable, []string{
			o.ResourceType.Id,
			o.ResourceType.DisplayName,
			strings.Join(traits, ", "),
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithData(resourceTypesTable).Render()
	if err != nil {
		return err
	}

	return nil
}

func (c *consoleManager) outputResources(out *v1.ResourceListOutput) error {
	resourcesTable := pterm.TableData{
		{"ID", "Display Name", "Resource Type", "Parent Resource"},
	}
	for _, r := range out.Resources {
		parentResourceText := "-"
		if r.Parent != nil {
			parentResourceText = fmt.Sprintf(
				"%s (%s)",
				r.Parent.DisplayName,
				r.Parent.Id.ResourceType,
			)
		}

		resourcesTable = append(resourcesTable, []string{
			r.Resource.Id.Resource,
			r.Resource.DisplayName,
			r.ResourceType.DisplayName,
			parentResourceText,
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
	if err != nil {
		return err
	}

	return nil
}

func (c *consoleManager) outputEntitlements(out *v1.EntitlementListOutput) error {
	entitlementsTable := pterm.TableData{
		{"ID", "Display Name", "Resource Type", "Resource", "Permission"},
	}
	for _, u := range out.Entitlements {
		entitlementsTable = append(entitlementsTable, []string{
			u.Entitlement.Id,
			u.Entitlement.DisplayName,
			u.ResourceType.DisplayName,
			u.Resource.DisplayName,
			u.Entitlement.Slug,
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithData(entitlementsTable).Render()
	if err != nil {
		return err
	}

	return nil
}

func (c *consoleManager) outputGrants(out *v1.GrantListOutput) error {
	grantsTable := pterm.TableData{
		{"ID", "Resource Type", "Resource", "Entitlement", "Principal"},
	}

	for _, g := range out.Grants {
		grantsTable = append(grantsTable, []string{
			g.Grant.Id,
			g.ResourceType.DisplayName,
			g.Resource.DisplayName,
			g.Entitlement.DisplayName,
			g.Principal.DisplayName,
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithData(grantsTable).Render()
	if err != nil {
		return err
	}

	return nil
}

func (c *consoleManager) outputResourceAccess(out *v1.ResourceAccessListOutput) error {
	leveledList := pterm.LeveledList{
		pterm.LeveledListItem{
			Level: 0,
			Text:  fmt.Sprintf("Effective Access for %s (%s)", out.Principal.DisplayName, out.Principal.Id.ResourceType),
		},
	}
	for _, g := range out.Access {
		leveledList = append(
			leveledList,
			pterm.LeveledListItem{Level: 1, Text: fmt.Sprintf("%s (%s)", g.Resource.DisplayName, g.ResourceType.DisplayName)},
		)

		for _, e := range g.Entitlements {
			leveledList = append(
				leveledList,
				pterm.LeveledListItem{Level: 2, Text: e.Slug},
			)
		}
	}

	root := putils.TreeFromLeveledList(leveledList)
	err := pterm.DefaultTree.WithRoot(root).Render()
	if err != nil {
		return err
	}

	return nil
}

func (c *consoleManager) outputPrincipalsCompare(out *v1.PrincipalsCompareOutput) error {
	if len(out.Missing) == 0 && len(out.Extra) == 0 {
		fmt.Fprintf(os.Stdout, "The principals between these entitlements appear to match!")
		return nil
	}

	if len(out.Missing) > 0 {
		fmt.Fprintf(os.Stdout, "\n")
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).Println("Missing Principals")
		fmt.Fprintf(os.Stdout, "\n")

		resourcesTable := pterm.TableData{
			{"ID", "Display Name", "Resource Type", "Parent Resource"},
		}

		for _, r := range out.Missing {
			parentResourceText := "-"
			if r.Parent != nil {
				parentResourceText = fmt.Sprintf(
					"%s (%s)",
					r.Parent.DisplayName,
					r.Parent.Id.ResourceType,
				)
			}

			resourcesTable = append(resourcesTable, []string{
				r.Resource.Id.Resource,
				r.Resource.DisplayName,
				r.ResourceType.DisplayName,
				parentResourceText,
			})
		}

		err := pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
		if err != nil {
			return err
		}
	}

	if len(out.Extra) > 0 {
		fmt.Fprintf(os.Stdout, "\n")
		pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgLightBlue)).Println("Extra Principals")
		fmt.Fprintf(os.Stdout, "\n")

		resourcesTable := pterm.TableData{
			{"ID", "Display Name", "Resource Type", "Parent Resource"},
		}

		for _, r := range out.Extra {
			parentResourceText := "-"
			if r.Parent != nil {
				parentResourceText = fmt.Sprintf(
					"%s (%s)",
					r.Parent.DisplayName,
					r.Parent.Id.ResourceType,
				)
			}

			resourcesTable = append(resourcesTable, []string{
				r.Resource.Id.Resource,
				r.Resource.DisplayName,
				r.ResourceType.DisplayName,
				parentResourceText,
			})
		}

		err := pterm.DefaultTable.WithHasHeader().WithData(resourcesTable).Render()
		if err != nil {
			return err
		}
	}

	return nil
}
