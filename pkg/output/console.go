package output

import (
	"context"
	"fmt"
	"strings"

	v1 "github.com/conductorone/baton-cli/pb/baton_cli/v1"
	"github.com/pterm/pterm"
)

type consoleManager struct{}

func (c *consoleManager) Output(ctx context.Context, out interface{}) error {
	switch obj := out.(type) {
	case *v1.ResourceTypeListOutput:
		return c.outputResourceTypes(obj)

	case *v1.ResourceListOutput:
		return c.outputResources(obj)

	default:
		return fmt.Errorf("unexpected output model")
	}
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
				"%s (%s - %s)",
				r.Parent.DisplayName,
				r.Parent.Id.ResourceType,
				r.Parent.Id.Resource,
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
