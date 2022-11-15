package manager

import (
	"context"
	"strings"

	"github.com/conductorone/baton-sdk/internal/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager/local"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager/s3"
)

type Manager interface {
	LoadC1Z(ctx context.Context) (*dotc1z.C1File, error)
	SaveC1Z(ctx context.Context) error
	Close(ctx context.Context) error
}

// Given a file path, return a Manager that can read and write files to that path.
//
// The first thing we do is check if the file path starts with "s3://". If it does, we return a new
// S3Manager. If it doesn't, we return a new LocalManager.
func New(ctx context.Context, filePath string) (Manager, error) {
	switch {
	case strings.HasPrefix(filePath, "s3://"):
		return s3.NewS3Manager(ctx, filePath)
	default:
		return local.New(ctx, filePath)
	}
}
