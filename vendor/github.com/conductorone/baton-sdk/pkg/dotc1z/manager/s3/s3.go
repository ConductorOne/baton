package s3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/aws/smithy-go"
	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/us3"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.uber.org/zap"
)

type s3Manager struct {
	client   *us3.S3Client
	fileName string
	tmpFile  string
	tmpDir   string
}

type Option func(*s3Manager)

func WithTmpDir(tmpDir string) Option {
	return func(o *s3Manager) {
		o.tmpDir = tmpDir
	}
}

func (s *s3Manager) copyToTempFile(ctx context.Context, r io.Reader) error {
	f, err := os.CreateTemp(s.tmpDir, "sync-*.c1z")
	if err != nil {
		return err
	}
	defer f.Close()

	s.tmpFile = f.Name()

	if r != nil {
		_, err = io.Copy(f, r)
		if err != nil {
			_ = f.Close()
			return err
		}
	}

	return nil
}

// LoadRaw loads the file from S3 and returns an io.Reader for the contents.
func (s *s3Manager) LoadRaw(ctx context.Context) (io.ReadCloser, error) {
	out, err := s.client.Get(ctx, s.fileName)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "NotFound":
				return nil, err
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	err = s.copyToTempFile(ctx, out)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(s.tmpFile)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// LoadC1Z gets a file from the AWS S3 bucket and copies it to a temp file.
func (s *s3Manager) LoadC1Z(ctx context.Context) (*dotc1z.C1File, error) {
	l := ctxzap.Extract(ctx)

	out, err := s.client.Get(ctx, s.fileName)
	if err != nil {
		var ae smithy.APIError
		if errors.As(err, &ae) {
			switch ae.ErrorCode() {
			case "NotFound":
				l.Info("c1z was not found in s3 -- creating empty c1z", zap.String("file_path", s.fileName))
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	err = s.copyToTempFile(ctx, out)
	if err != nil {
		return nil, err
	}

	return dotc1z.NewC1ZFile(ctx, s.tmpFile, dotc1z.WithTmpDir(s.tmpDir))
}

// SaveC1Z saves a file to the AWS S3 bucket.
func (s *s3Manager) SaveC1Z(ctx context.Context) error {
	f, err := os.Open(s.tmpFile)
	if err != nil {
		return err
	}

	if s.client == nil {
		return fmt.Errorf("attempting to save to s3 without a valid client")
	}

	if s.fileName == "" {
		return fmt.Errorf("attempting to save to s3 without a valid file path specified")
	}

	err = s.client.Put(ctx, s.fileName, f, "application/c1z")
	if err != nil {
		return err
	}

	return nil
}

func (s *s3Manager) Close(ctx context.Context) error {
	err := os.Remove(s.tmpFile)
	if err != nil {
		return err
	}

	return nil
}

// NewS3Manager returns a new `s3Manager` that uses the given `s3Uri`.
func NewS3Manager(ctx context.Context, s3Uri string, opts ...Option) (*s3Manager, error) {
	l := ctxzap.Extract(ctx)

	fileName, s3Client, err := us3.NewClientFromURI(ctx, s3Uri)
	if err != nil {
		return nil, err
	}

	manager := &s3Manager{
		client:   s3Client,
		fileName: fileName,
	}

	for _, opt := range opts {
		opt(manager)
	}

	l.Debug("created new s3 file manager", zap.String("filename", fileName))

	return manager, nil
}
