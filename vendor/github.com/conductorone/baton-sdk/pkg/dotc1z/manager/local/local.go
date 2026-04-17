package local

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"

	"github.com/conductorone/baton-sdk/pkg/dotc1z"
	"github.com/conductorone/baton-sdk/pkg/uotel"
)

var tracer = otel.Tracer("baton-sdk/pkg.dotc1z.manager.local")

type localManager struct {
	filePath       string
	tmpPath        string
	tmpDir         string
	decoderOptions []dotc1z.DecoderOption
	skipCleanup    bool
}

type Option func(*localManager)

func WithTmpDir(tmpDir string) Option {
	return func(o *localManager) {
		o.tmpDir = tmpDir
	}
}

func WithDecoderOptions(opts ...dotc1z.DecoderOption) Option {
	return func(o *localManager) {
		o.decoderOptions = opts
	}
}

func WithSkipCleanup(skip bool) Option {
	return func(o *localManager) {
		o.skipCleanup = skip
	}
}

func (l *localManager) copyFileToTmp(ctx context.Context) error {
	_, span := tracer.Start(ctx, "localManager.copyFileToTmp")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	tmp, err := os.CreateTemp(l.tmpDir, "sync-*.c1z")
	if err != nil {
		return err
	}
	defer tmp.Close()

	l.tmpPath = tmp.Name()

	if l.filePath == "" {
		return nil
	}

	if _, err = os.Stat(l.filePath); err == nil {
		f, err := os.Open(l.filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		_, err = io.Copy(tmp, f)
		if err != nil {
			return err
		}

		if err := tmp.Sync(); err != nil {
			return fmt.Errorf("failed to sync temp file: %w", err)
		}
	}

	return nil
}

// LoadRaw returns an io.Reader of the bytes in the c1z file.
func (l *localManager) LoadRaw(ctx context.Context) (io.ReadCloser, error) {
	ctx, span := tracer.Start(ctx, "localManager.LoadRaw")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	err = l.copyFileToTmp(ctx)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(l.tmpPath)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// LoadC1Z loads the C1Z file from the local file system.
func (l *localManager) LoadC1Z(ctx context.Context) (*dotc1z.C1File, error) {
	ctx, span := tracer.Start(ctx, "localManager.LoadC1Z")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	log := ctxzap.Extract(ctx)

	err = l.copyFileToTmp(ctx)
	if err != nil {
		return nil, err
	}

	log.Debug(
		"successfully loaded c1z locally",
		zap.String("file_path", l.filePath),
		zap.String("temp_path", l.tmpPath),
		zap.String("tmp_dir", l.tmpDir),
	)

	opts := []dotc1z.C1ZOption{
		dotc1z.WithTmpDir(l.tmpDir),
	}
	if len(l.decoderOptions) > 0 {
		opts = append(opts, dotc1z.WithDecoderOptions(l.decoderOptions...))
	}
	if l.skipCleanup {
		opts = append(opts, dotc1z.WithSkipCleanup(true))
	}
	return dotc1z.NewC1ZFile(ctx, l.tmpPath, opts...)
}

// SaveC1Z saves the C1Z file to the local file system.
func (l *localManager) SaveC1Z(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "localManager.SaveC1Z")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	log := ctxzap.Extract(ctx)

	if l.tmpPath == "" {
		return fmt.Errorf("unexpected state - missing temp file path")
	}

	if l.filePath == "" {
		return fmt.Errorf("unexpected state - missing file path")
	}

	tmpFile, err := os.Open(l.tmpPath)
	if err != nil {
		return err
	}
	defer tmpFile.Close()

	dstFile, err := os.Create(l.filePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	size, err := io.Copy(dstFile, tmpFile)
	if err != nil {
		return err
	}

	// CRITICAL: Sync to ensure data is written before function returns.
	// This is especially important on ZFS ARC where writes may be cached.
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync destination file: %w", err)
	}

	log.Debug(
		"successfully saved c1z locally",
		zap.String("file_path", l.filePath),
		zap.String("temp_path", l.tmpPath),
		zap.String("tmp_dir", l.tmpDir),
		zap.Int64("bytes", size),
	)

	return nil
}

func (l *localManager) Close(ctx context.Context) error {
	_, span := tracer.Start(ctx, "localManager.Close")
	var err error
	defer func() { uotel.EndSpanWithError(span, err) }()

	err = os.Remove(l.tmpPath)
	if err != nil {
		return err
	}
	return nil
}

// New returns a new localManager that uses the given filePath.
func New(ctx context.Context, filePath string, opts ...Option) (*localManager, error) {
	ret := &localManager{
		filePath: filePath,
	}

	for _, opt := range opts {
		opt(ret)
	}

	return ret, nil
}
