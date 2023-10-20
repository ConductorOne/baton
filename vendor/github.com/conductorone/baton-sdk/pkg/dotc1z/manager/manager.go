package manager

import (
	"context"
	"io"
	"strings"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager/local"
	"github.com/conductorone/baton-sdk/pkg/dotc1z/manager/s3"
)

type Manager interface {
	LoadRaw(ctx context.Context) (io.ReadCloser, error)
	LoadC1Z(ctx context.Context) (*dotc1z.C1File, error)
	SaveC1Z(ctx context.Context) error
	Close(ctx context.Context) error
}

type managerOptions struct {
	tmpDir string
}

type ManagerOption func(*managerOptions)

func WithTmpDir(tmpDir string) ManagerOption {
	return func(o *managerOptions) {
		o.tmpDir = tmpDir
	}
}

// Given a file path, return a Manager that can read and write files to that path.
//
// The first thing we do is check if the file path starts with "s3://". If it does, we return a new
// S3Manager. If it doesn't, we return a new LocalManager.
func New(ctx context.Context, filePath string, opts ...ManagerOption) (Manager, error) {
	options := &managerOptions{}

	for _, opt := range opts {
		opt(options)
	}

	switch {
	case strings.HasPrefix(filePath, "s3://"):
		var s3Opts []s3.Option
		if options.tmpDir != "" {
			s3Opts = append(s3Opts, s3.WithTmpDir(options.tmpDir))
		}
		return s3.NewS3Manager(ctx, filePath, s3Opts...)
	default:
		var localOpts []local.Option
		if options.tmpDir != "" {
			localOpts = append(localOpts, local.WithTmpDir(options.tmpDir))
		}
		return local.New(ctx, filePath, localOpts...)
	}
}
