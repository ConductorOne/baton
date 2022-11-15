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
	case *v1.ResourceTypeOutput:
		return c.outputResourceTypes(obj)

	default:
		return fmt.Errorf("unexpected output model")
	}
}

func (c *consoleManager) outputResourceTypes(out *v1.ResourceTypeOutput) error {
	resourceTypesTable := pterm.TableData{
		{"ID", "Display Name", "Traits"},
	}

	for _, rt := range out.ResourceTypes {
		var traits []string
		for _, t := range rt.Traits {
			traits = append(traits, t.String())
		}

		resourceTypesTable = append(resourceTypesTable, []string{
			rt.Id,
			rt.DisplayName,
			strings.Join(traits, ", "),
		})
	}

	err := pterm.DefaultTable.WithHasHeader().WithData(resourceTypesTable).Render()
	if err != nil {
		return err
	}

	return nil
}
