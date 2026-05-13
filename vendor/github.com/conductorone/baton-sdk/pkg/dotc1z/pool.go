package dotc1z

import (
	"os"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// poolDisabled is an operational kill-switch. Set BATON_ZSTD_POOL_DISABLE=1
// in the environment to bypass the encoder and decoder pools entirely and
// let callers fall back to fresh encoder/decoder creation. Zero behavior
// change when unset. Useful for rolling back pool behavior without a
// revert + re-release cycle if a latent Reset or Close bug surfaces in prod.
var poolDisabled = os.Getenv("BATON_ZSTD_POOL_DISABLE") == "1"

// encoderPool manages reusable zstd.Encoder instances to reduce allocation overhead.
// Pooled encoders are single-threaded (concurrency=1) to match the production
// default set in c1file.go. A concurrent encoder path (GOMAXPROCS) is intentionally
// avoided here because klauspost/compress's multi-threaded encoder has theoretical
// race concerns in its goroutine synchronization.
var encoderPool = &sync.Pool{}

// pooledEncoderConcurrency is the concurrency level used for pooled encoders.
// Must match the c1zOptions / C1File default (1) so the pool actually engages
// on the sync hot path. Note: pkg/synccompactor/compactor.go passes
// WithEncoderConcurrency(0) (GOMAXPROCS) and intentionally bypasses the pool.
const pooledEncoderConcurrency = 1

// getEncoder retrieves a zstd encoder from the pool or creates a new one.
// The returned encoder is NOT bound to any writer - call Reset(w) before use.
// Returns the encoder and a boolean indicating if it came from the pool.
func getEncoder() (*zstd.Encoder, bool) {
	if poolDisabled {
		return nil, false
	}
	if enc, ok := encoderPool.Get().(*zstd.Encoder); ok && enc != nil {
		return enc, true
	}

	// Create new encoder with default concurrency.
	// This should not fail with valid options, but handle it gracefully.
	enc, err := zstd.NewWriter(nil,
		zstd.WithEncoderConcurrency(pooledEncoderConcurrency),
	)
	if err != nil {
		// Fallback: return nil and let caller create encoder with their options
		return nil, false
	}
	return enc, false
}

// putEncoder returns a zstd encoder to the pool for reuse.
// The encoder is reset to release any reference to the previous writer.
// Encoders should be in a clean state (Close() called) before returning.
func putEncoder(enc *zstd.Encoder) {
	if enc == nil {
		return
	}
	if poolDisabled {
		return
	}
	// Reset to nil writer to release reference to previous output.
	// This is safe even if the encoder was already closed.
	enc.Reset(nil)
	encoderPool.Put(enc)
}

// decoderPool manages reusable zstd.Decoder instances to reduce allocation overhead.
// All pooled decoders are configured with concurrency=1 (single-threaded) and low memory mode.
var decoderPool = &sync.Pool{}

// getDecoder retrieves a zstd decoder from the pool or creates a new one.
// The returned decoder is NOT bound to any reader - call Reset(r) before use.
// Returns the decoder and a boolean indicating if it came from the pool.
func getDecoder() (*zstd.Decoder, bool) {
	if poolDisabled {
		return nil, false
	}
	if dec, ok := decoderPool.Get().(*zstd.Decoder); ok && dec != nil {
		return dec, true
	}

	// Create new decoder with default options matching decoder.go defaults.
	dec, err := zstd.NewReader(nil,
		zstd.WithDecoderConcurrency(1),
		zstd.WithDecoderLowmem(true),
		zstd.WithDecoderMaxMemory(defaultDecoderMaxMemory),
	)
	if err != nil {
		// Fallback: return nil and let caller create decoder with their options
		return nil, false
	}
	return dec, false
}

// putDecoder returns a zstd decoder to the pool for reuse.
// The decoder is reset to release any reference to the previous reader.
func putDecoder(dec *zstd.Decoder) {
	if dec == nil {
		return
	}
	if poolDisabled {
		return
	}
	// Reset to nil reader to release reference to previous input.
	// If Reset fails (bad state), don't return to pool.
	if err := dec.Reset(nil); err != nil {
		return
	}
	decoderPool.Put(dec)
}
