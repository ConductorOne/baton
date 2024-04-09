package dotc1z

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"sync"

	"github.com/klauspost/compress/zstd"
)

const (
	defaultMaxDecodedSize   = 2 * 1024 * 1024 * 1024 // 2GiB
	defaultDecoderMaxMemory = 32 * 1024 * 1024       // 32MiB
	maxDecodedSizeEnvVar    = "BATON_DECODER_MAX_DECODED_SIZE_MB"
	maxDecoderMemorySizeEnv = "BATON_DECODER_MAX_MEMORY_MB"
)

var C1ZFileHeader = []byte("C1ZF\x00")

var (
	ErrInvalidFile        = fmt.Errorf("c1z: invalid file")
	ErrMaxSizeExceeded    = errors.New("c1z: max decoded size exceeded, increase DecoderMaxDecodedSize")
	ErrWindowSizeExceeded = errors.New("c1z: window size exceeded, increase DecoderMaxMemory")
)

// ReadHeader reads len(C1ZFileHeader) bytes from the given io.Reader and compares them to C1ZFileHeader, returning an error if they don't match.
// If possible, ReadHeader will Seek() to the start of the stream before checking the header bytes.
// On return, the reader will be pointing to the first byte after the header.
func ReadHeader(reader io.Reader) error {
	rs, ok := reader.(io.Seeker)
	if ok {
		_, err := rs.Seek(0, 0)
		if err != nil {
			return err
		}
	}

	headerBytes := make([]byte, len(C1ZFileHeader))
	_, err := reader.Read(headerBytes)
	if err != nil {
		return err
	}

	if !bytes.Equal(headerBytes, C1ZFileHeader) {
		return ErrInvalidFile
	}

	return nil
}

// DecoderOption is an option for creating a decoder.
type DecoderOption func(*decoderOptions) error

// options retains accumulated state of multiple options.
type decoderOptions struct {
	ctx            context.Context
	maxDecodedSize uint64
	maxMemorySize  uint64
}

// WithContext sets a context, when cancelled, will cause subequent calls to Read() to return ctx.Error().
func WithContext(ctx context.Context) DecoderOption {
	return func(o *decoderOptions) error {
		o.ctx = ctx
		return nil
	}
}

// WithDecoderMaxMemory sets the maximum window size for streaming operations.
// This can be used to control memory usage of potentially hostile content.
// Maximum is 1 << 63 bytes. Default is 32MiB.
func WithDecoderMaxMemory(n uint64) DecoderOption {
	return func(o *decoderOptions) error {
		if n == 0 {
			return errors.New("c1z: WithDecoderMaxMemory must be at least 1")
		}
		if n > 1<<63 {
			return errors.New("c1z: WithDecoderMaxMemory must be less than 1 << 63")
		}
		o.maxMemorySize = n
		return nil
	}
}

// WithDecoderMaxDecodedSize sets the maximum size of the decoded stream.
// This can be used to cap the resulting decoded stream size.
// Maximum is 1 << 63 bytes. Default is 1GiB.
func WithDecoderMaxDecodedSize(n uint64) DecoderOption {
	return func(o *decoderOptions) error {
		if n == 0 {
			return errors.New("c1z: WithDecoderMaxDecodedSize must be at least 1")
		}
		if n > 1<<63 {
			return errors.New("c1z: WithDecoderMaxDecodedSize must be less than 1 << 63")
		}
		o.maxDecodedSize = n
		return nil
	}
}

type decoder struct {
	o  *decoderOptions
	f  io.Reader
	zd *zstd.Decoder

	decodedBytes uint64

	initOnce       sync.Once
	headerCheckErr error
	decoderInitErr error
}

func (d *decoder) Read(p []byte) (int, error) {
	// Init
	d.initOnce.Do(func() {
		err := ReadHeader(d.f)
		if err != nil {
			d.headerCheckErr = err
			return
		}

		maxMemSize := d.o.maxMemorySize
		if maxMemSize == 0 {
			maxMemSize = defaultDecoderMaxMemory
		}
		zd, err := zstd.NewReader(
			d.f,
			zstd.WithDecoderConcurrency(1),        // disables async decoding/concurrency
			zstd.WithDecoderLowmem(true),          // uses lower memory, trading potentially more allocations
			zstd.WithDecoderMaxMemory(maxMemSize), // sets limit on maximum memory used when decoding stream
		)
		if err != nil {
			d.decoderInitErr = err
			return
		}
		d.zd = zd
	})

	// Check header
	if d.headerCheckErr != nil {
		return 0, d.headerCheckErr
	}

	// Check we have a valid decoder
	if d.zd != nil && d.decoderInitErr != nil {
		return 0, d.decoderInitErr
	}

	// Check our (optional) context is not cancelled
	if d.o.ctx != nil && d.o.ctx.Err() != nil {
		return 0, d.o.ctx.Err()
	}

	// Check we have not exceeded our max decoded size
	maxDecodedSize := d.o.maxDecodedSize
	if maxDecodedSize == 0 {
		maxDecodedSize = defaultMaxDecodedSize
	}
	if d.decodedBytes > maxDecodedSize {
		return 0, ErrMaxSizeExceeded
	}

	// Do underlying read
	n, err := d.zd.Read(p)
	d.decodedBytes += uint64(n)
	if err != nil {
		// NOTE(morgabra) This happens if you set a small DecoderMaxMemory
		if errors.Is(err, zstd.ErrWindowSizeExceeded) {
			return n, ErrWindowSizeExceeded
		}
		return n, err
	}
	return n, nil
}

func (d *decoder) Close() error {
	if d.zd != nil {
		d.zd.Close()
	}
	return nil
}

// NewDecoder wraps a given .c1z file io.Reader and returns an io.Reader for the underlying decoded/uncompressed file.
func NewDecoder(f io.Reader, opts ...DecoderOption) (*decoder, error) {
	// We want these options to be configurable via the environment. They are appended to the end of opts so they will take
	// precedence over any other options of the same type.
	maxDecodedSizeVar := os.Getenv(maxDecodedSizeEnvVar)
	if maxDecodedSizeVar != "" {
		maxDecodedSize, err := strconv.ParseUint(maxDecodedSizeVar, 10, 64)
		if err == nil {
			opts = append(opts, WithDecoderMaxDecodedSize(maxDecodedSize*1024*1024))
		}
	}

	maxDecoderMemorySizeVar := os.Getenv(maxDecoderMemorySizeEnv)
	if maxDecoderMemorySizeVar != "" {
		maxDecoderMemorySize, err := strconv.ParseUint(maxDecoderMemorySizeVar, 10, 64)
		if err == nil {
			opts = append(opts, WithDecoderMaxMemory(maxDecoderMemorySize*1024*1024))
		}
	}

	o := &decoderOptions{}
	for _, opt := range opts {
		err := opt(o)
		if err != nil {
			return nil, err
		}
	}

	return &decoder{
		o: o,
		f: f,
	}, nil
}
