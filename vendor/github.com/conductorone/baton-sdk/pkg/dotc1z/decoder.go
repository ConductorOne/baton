package dotc1z

import (
	"bufio"
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
	defaultMaxDecodedSize   = 10 * 1024 * 1024 * 1024 // 10GiB
	defaultDecoderMaxMemory = 128 * 1024 * 1024       // 128MiB
	maxDecodedSizeEnvVar    = "BATON_DECODER_MAX_DECODED_SIZE_MB"
	maxDecoderMemorySizeEnv = "BATON_DECODER_MAX_MEMORY_MB"

	// fcsFailFastDisableEnvVar, when set to "1", disables the fail-fast path
	// that rejects c1zs whose declared Frame_Content_Size already exceeds the
	// configured DecoderMaxDecodedSize. Disabling falls back to the pre-FCS
	// behavior of running the decoder until decoded bytes exceed the cap.
	// Intended as an operational kill-switch if an FCS-induced false positive
	// surfaces in production; zero behavior change when unset.
	fcsFailFastDisableEnvVar = "BATON_DISABLE_FCS_FAIL_FAST"
)

// fcsFailFastDisabled is read once at package init to avoid per-Read env
// lookups. Matches the pattern used by BATON_ZSTD_POOL_DISABLE in pool.go.
var fcsFailFastDisabled = os.Getenv(fcsFailFastDisableEnvVar) == "1"

var C1ZFileHeader = []byte("C1ZF\x00")

var (
	ErrInvalidFile        = fmt.Errorf("c1z: invalid file")
	ErrMaxSizeExceeded    = fmt.Errorf("c1z: max decoded size exceeded, increase DecoderMaxDecodedSize using %v environment variable", maxDecodedSizeEnvVar)
	ErrWindowSizeExceeded = fmt.Errorf("c1z: window size exceeded, increase DecoderMaxMemory using %v  environment variable", maxDecoderMemorySizeEnv)
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
	ctx                context.Context
	maxDecodedSize     uint64
	maxMemorySize      uint64
	decoderConcurrency int
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
// Maximum is 1 << 63 bytes. Default is 128MiB.
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

// WithDecoderConcurrency sets the number of created decoders.
// Default is 1, which disables async decoding/concurrency.
// 0 uses GOMAXPROCS.
// -1 uses GOMAXPROCS or 4, whichever is lower.
func WithDecoderConcurrency(n int) DecoderOption {
	return func(o *decoderOptions) error {
		o.decoderConcurrency = n
		return nil
	}
}

type decoder struct {
	o  *decoderOptions
	f  io.Reader
	zd *zstd.Decoder

	decodedBytes   uint64
	poolCompatible bool // true if zd has pool-compatible settings and should be returned to pool

	// declaredSize is the Frame_Content_Size from the zstd frame header, if
	// advertised. Populated lazily during init (alongside header validation),
	// before any decompression happens.
	declaredSize    uint64
	hasDeclaredSize bool

	initOnce       sync.Once
	headerCheckErr error
	decoderInitErr error
}

func (d *decoder) getMaxMemSize() uint64 {
	maxMemSize := d.o.maxMemorySize
	if maxMemSize == 0 {
		maxMemSize = defaultDecoderMaxMemory
	}
	return maxMemSize
}

func (d *decoder) getMaxDecodedSize() uint64 {
	v := d.o.maxDecodedSize
	if v == 0 {
		v = defaultMaxDecodedSize
	}
	return v
}

// DeclaredDecodedSize returns the decompressed size advertised in the zstd
// Frame_Content_Size field of the c1z's frame header, if present. It triggers
// header parsing on first call without starting decompression.
//
// Returns (size, true) when FCS is advertised — this is the exact number of
// bytes the stream will produce. Returns (0, false) for c1zs saved before the
// producer started recording FCS, or when header parsing failed (in which case
// Read will surface the underlying error).
//
// Safe to call before, during, or instead of Read; subsequent Read calls
// continue to work normally.
func (d *decoder) DeclaredDecodedSize() (uint64, bool) {
	d.ensureInit()
	return d.declaredSize, d.hasDeclaredSize
}

// ensureInit runs decoder initialization at most once: header check, FCS peek,
// fail-fast on declared-size cap, then zstd decoder setup (from pool or fresh).
// Called from both Read and DeclaredDecodedSize.
func (d *decoder) ensureInit() {
	d.initOnce.Do(func() {
		err := ReadHeader(d.f)
		if err != nil {
			d.headerCheckErr = err
			return
		}

		// Wrap the post-magic reader so we can peek the zstd frame header
		// without consuming those bytes — the zstd decoder still reads them
		// from the bufio buffer when it starts.
		br := bufio.NewReader(d.f)
		if peek, perr := br.Peek(zstd.HeaderMaxSize); perr == nil || errors.Is(perr, io.EOF) {
			var hdr zstd.Header
			if decErr := hdr.Decode(peek); decErr == nil && hdr.HasFCS {
				d.declaredSize = hdr.FrameContentSize
				d.hasDeclaredSize = true
			}
		}

		// Fail fast when the declared size already exceeds the caller's cap:
		// no need to instantiate a zstd decoder for a c1z we can't accept.
		// The error wraps ErrMaxSizeExceeded and carries the declared size
		// explicitly so callers can surface "this c1z wants to decode to N
		// bytes" without running the decoder. BATON_DISABLE_FCS_FAIL_FAST=1
		// skips the pre-check and falls back to the decoded-bytes-over-cap
		// path inside Read (pre-FCS behavior).
		if d.hasDeclaredSize && !fcsFailFastDisabled {
			maxDecodedSize := d.getMaxDecodedSize()
			if d.declaredSize > maxDecodedSize {
				d.decoderInitErr = fmt.Errorf(
					"c1z: declared decoded size %d exceeds cap %d: %w",
					d.declaredSize, maxDecodedSize, ErrMaxSizeExceeded,
				)
				return
			}
		}

		// Try to use a pooled decoder if options match the pool's defaults.
		// Pool decoders use: concurrency=1, lowmem=true, maxMemory=defaultDecoderMaxMemory.
		usePool := d.o.decoderConcurrency == 1 && d.getMaxMemSize() == defaultDecoderMaxMemory
		if usePool {
			zd, _ := getDecoder()
			if zd != nil {
				if err := zd.Reset(br); err != nil {
					// Reset failed, return decoder to pool and fall through to create new one.
					putDecoder(zd)
				} else {
					d.zd = zd
					d.poolCompatible = true // Mark for return to pool on Close()
					return
				}
			}
		}

		// Non-default options or pool unavailable: create new decoder.
		zstdOpts := []zstd.DOption{
			zstd.WithDecoderLowmem(true),                 // uses lower memory, trading potentially more allocations
			zstd.WithDecoderMaxMemory(d.getMaxMemSize()), // sets limit on maximum memory used when decoding stream
		}
		if d.o.decoderConcurrency >= 0 {
			zstdOpts = append(zstdOpts, zstd.WithDecoderConcurrency(d.o.decoderConcurrency))
		}

		zd, err := zstd.NewReader(br, zstdOpts...)
		if err != nil {
			d.decoderInitErr = err
			return
		}
		d.zd = zd
		// If settings are pool-compatible, mark for return to pool on Close()
		d.poolCompatible = usePool
	})
}

func (d *decoder) Read(p []byte) (int, error) {
	d.ensureInit()

	// Check header
	if d.headerCheckErr != nil {
		return 0, d.headerCheckErr
	}

	// Check we have a valid decoder
	if d.decoderInitErr != nil {
		return 0, d.decoderInitErr
	}

	// Check our (optional) context is not cancelled
	if d.o.ctx != nil && d.o.ctx.Err() != nil {
		return 0, d.o.ctx.Err()
	}

	// Enforce max decoded size at both ends of the underlying Read:
	// - Top-of-call: short-circuit subsequent Reads after we've already tripped.
	// - Post-call: clip `n` so bytes beyond the cap never reach the caller,
	//   even when a single zstd read straddles the cap boundary.
	// When FCS is present and within the cap the post-call branch is
	// unreachable; it's defense against a missing or mismatched FCS.
	maxDecodedSize := d.getMaxDecodedSize()
	if d.decodedBytes > maxDecodedSize {
		return 0, d.maxSizeExceededErr()
	}

	// Do underlying read
	n, err := d.zd.Read(p)
	//nolint:gosec // No risk of overflow/underflow because n is always >= 0.
	d.decodedBytes += uint64(n)

	// Clip any bytes that crossed the cap in this single Read so we never
	// return data past maxDecodedSize, and surface the error immediately
	// instead of waiting for the next Read.
	if d.decodedBytes > maxDecodedSize {
		overrun := d.decodedBytes - maxDecodedSize
		//nolint:gosec // overrun <= n by construction (we just added n and went over).
		if uint64(n) >= overrun {
			n -= int(overrun)
		} else {
			n = 0
		}
		// Leave d.decodedBytes at its true post-read value so subsequent
		// Reads hit the top-of-call guard above.
		return n, d.maxSizeExceededErr()
	}

	if err != nil {
		// NOTE(morgabra) This happens if you set a small DecoderMaxMemory
		if errors.Is(err, zstd.ErrWindowSizeExceeded) {
			return n, fmt.Errorf("c1z: window size (%d) exceeded:  %w: %w", d.getMaxMemSize(), err, ErrWindowSizeExceeded)
		}
		return n, err
	}
	return n, nil
}

// maxSizeExceededErr builds the ErrMaxSizeExceeded wrap emitted from both the
// top-of-Read guard and the in-Read clip path. Includes the declared size
// when FCS was advertised so callers see the exact stream size without
// decompressing further.
func (d *decoder) maxSizeExceededErr() error {
	maxDecodedSize := d.getMaxDecodedSize()
	if d.hasDeclaredSize {
		return fmt.Errorf(
			"c1z: max decoded size exceeded: %d > %d (declared: %d): %w",
			d.decodedBytes, maxDecodedSize, d.declaredSize, ErrMaxSizeExceeded,
		)
	}
	return fmt.Errorf("c1z: max decoded size exceeded: %d > %d: %w", d.decodedBytes, maxDecodedSize, ErrMaxSizeExceeded)
}

func (d *decoder) Close() error {
	if d.zd != nil {
		if d.poolCompatible {
			// Return decoder to pool for reuse.
			putDecoder(d.zd)
		} else {
			d.zd.Close()
		}
		d.zd = nil
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

	o := &decoderOptions{
		decoderConcurrency: 1,
	}
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
