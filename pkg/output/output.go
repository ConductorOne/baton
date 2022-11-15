package output

import (
	"context"
)

type Manager interface {
	Output(ctx context.Context, out interface{}) error
}

func NewManager(ctx context.Context, format string) Manager {
	switch format {
	case "console":
		return &consoleManager{}
	case "json":
		return &jsonManager{}
	default:
		return &consoleManager{}
	}
}
