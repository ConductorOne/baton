package dotc1z

import (
	"context"
	"errors"
	"fmt"
	"io"

	"go.opentelemetry.io/otel"

	"github.com/conductorone/baton-sdk/pkg/connectorstore"
)

var tracer = otel.Tracer("baton-sdk/pkg.dotc1z")

// NewC1FileReader returns a connectorstore.Reader implementation for the given sqlite db file path.
func NewC1FileReader(ctx context.Context, dbFilePath string, opts ...C1FOption) (connectorstore.Reader, error) {
	return NewC1File(ctx, dbFilePath, opts...)
}

// NewC1ZFileDecoder wraps a given .c1z io.Reader that validates the .c1z and decompresses/decodes the underlying file.
// Defaults: 128MiB max memory and 3GiB max decoded size
// You must close the resulting io.ReadCloser when you are done, do not forget to close the given io.Reader if necessary.
func NewC1ZFileDecoder(f io.Reader, opts ...DecoderOption) (io.ReadCloser, error) {
	return NewDecoder(f, opts...)
}

// C1ZFileCheckHeader reads len(C1ZFileHeader) bytes from the given io.ReadSeeker and compares them to C1ZFileHeader.
// Returns true if the header is valid. Returns any errors from Read() or Seek().
// If a nil error is returned, the given io.ReadSeeker will be pointing to the first byte of the stream, and is suitable
// to be passed to NewC1ZFileDecoder.
func C1ZFileCheckHeader(f io.ReadSeeker) (bool, error) {
	// Read header
	err := ReadHeader(f)

	// Seek back to start
	_, seekErr := f.Seek(0, 0)
	if seekErr != nil {
		return false, err
	}

	if err != nil {
		if errors.Is(err, ErrInvalidFile) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func NewExternalC1FileReader(ctx context.Context, tmpDir string, externalResourceC1ZPath string) (connectorstore.Reader, error) {
	dbFilePath, _, err := decompressC1z(externalResourceC1ZPath, tmpDir)
	if err != nil {
		return nil, fmt.Errorf("error loading external resource c1z file: %w", err)
	}

	return NewC1File(ctx, dbFilePath)
}
