package dotc1z

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/klauspost/compress/zstd"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Note(kans): decompressC1z is unfortunately called to load or create a c1z file so the error handling is rough.
// It creates its own temporary directory so that it can also do its own cleanup.
// It returns that directory for verification in tests.
func decompressC1z(c1zPath string, workingDir string, opts ...DecoderOption) (string, string, error) {
	tmpDir, err := os.MkdirTemp(workingDir, "c1z")
	if err != nil {
		return "", tmpDir, err
	}

	var dbFile *os.File
	var c1zFile *os.File
	var decoder *decoder
	cleanupDir := func(e error) error {
		if decoder != nil {
			err := decoder.Close()
			if err != nil {
				e = errors.Join(e, err)
			}
		}
		if c1zFile != nil {
			err := c1zFile.Close()
			if err != nil {
				e = errors.Join(e, err)
			}
		}
		if dbFile != nil {
			err := dbFile.Close()
			if err != nil {
				e = errors.Join(e, err)
			}
		}
		if e != nil {
			err := os.RemoveAll(tmpDir)
			if err != nil {
				e = errors.Join(e, err)
			}
		}
		return e
	}

	dbFilePath := filepath.Join(tmpDir, "db")
	dbFile, err = os.Create(dbFilePath)
	if err != nil {
		return "", tmpDir, cleanupDir(err)
	}

	// #nosec G703 -- c1zPath is an explicit caller-provided local file path by design.
	stat, err := os.Stat(c1zPath)
	if err != nil || stat.Size() == 0 {
		// TODO(kans): it would be nice to know more about the error....
		return dbFilePath, tmpDir, cleanupDir(nil)
	}

	// #nosec G703 -- opening the caller-provided c1z file is the expected behavior.
	c1zFile, err = os.Open(c1zPath)
	if err != nil {
		return "", tmpDir, cleanupDir(err)
	}

	decoder, err = NewDecoder(c1zFile, opts...)
	if err != nil {
		return "", tmpDir, cleanupDir(err)
	}

	_, err = io.Copy(dbFile, decoder)
	if err != nil {
		return "", tmpDir, cleanupDir(err)
	}

	// CRITICAL: Sync the database file before returning to ensure all
	// decompressed data is on disk. On filesystems with aggressive caching
	// (like ZFS with large ARC), SQLite might otherwise open the file and
	// see incomplete data still in kernel buffers.
	err = dbFile.Sync()
	if err != nil {
		return "", tmpDir, cleanupDir(err)
	}

	return dbFilePath, tmpDir, cleanupDir(nil)
}

func saveC1z(dbFilePath string, outputFilePath string, encoderConcurrency int) error {
	if outputFilePath == "" {
		return status.Errorf(codes.InvalidArgument, "c1z: output file path not configured")
	}

	dbFile, err := os.Open(dbFilePath)
	if err != nil {
		return err
	}
	defer func() {
		if dbFile != nil {
			err = dbFile.Close()
			if err != nil {
				zap.L().Error("failed to close db file", zap.Error(err))
			}
		}
	}()

	// Capture the decompressed db file size up front. It is used for two
	// purposes:
	//   1. Embedded into the zstd frame header as Frame_Content_Size (FCS),
	//      making the decoded size discoverable to any zstd reader without
	//      having to decompress the stream.
	//   2. Logged alongside the resulting compressed size so operators can
	//      observe the decompressed size of every c1z we produce.
	dbStat, err := dbFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat db file: %w", err)
	}
	dbSize := dbStat.Size()

	// Write to a temporary file first to ensure atomic writes.
	// This prevents file corruption if the process crashes mid-write,
	// since the original file remains intact until the rename succeeds.
	tmpPath := outputFilePath + ".tmp"
	outFile, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// Track whether we successfully completed the write
	success := false
	defer func() {
		if outFile != nil {
			err = outFile.Close()
			if err != nil {
				zap.L().Error("failed to close out file", zap.Error(err))
			}
		}
		// Clean up temp file if we didn't successfully rename it
		if !success {
			if removeErr := os.Remove(tmpPath); removeErr != nil && !os.IsNotExist(removeErr) {
				zap.L().Error("failed to remove temp file", zap.Error(removeErr))
			}
		}
	}()

	// Write the magic file header
	_, err = outFile.Write(C1ZFileHeader)
	if err != nil {
		return err
	}

	// zstd.WithEncoderConcurrency does not work the same as WithDecoderConcurrency.
	// WithDecoderConcurrency uses GOMAXPROCS if set to 0.
	// WithEncoderConcurrency errors if set to 0 (but defaults to GOMAXPROCS).
	if encoderConcurrency == 0 {
		encoderConcurrency = runtime.GOMAXPROCS(0)
	}

	// Try to use a pooled encoder if concurrency matches the pool's default.
	// This reduces allocation overhead for the common case.
	var c1z *zstd.Encoder
	if encoderConcurrency == pooledEncoderConcurrency {
		c1z, _ = getEncoder()
	}
	if c1z == nil {
		// Non-default concurrency or pool returned nil: create new encoder.
		// The writer passed here is replaced by ResetContentSize below, so
		// io.Discard is only used as a placeholder during construction.
		var err error
		c1z, err = zstd.NewWriter(io.Discard,
			zstd.WithEncoderConcurrency(encoderConcurrency),
			zstd.WithEncoderLevel(zstd.SpeedFastest),
		)
		if err != nil {
			return err
		}
	}

	// Reset the encoder onto outFile and record the decompressed size in the
	// zstd Frame_Content_Size field. This is upstream-spec-compatible — old
	// decoders ignore FCS and continue to stream as today. New decoders (and
	// the zstd CLI) can observe the decoded size from the frame header
	// without having to decompress the stream.
	//
	// Note: ResetContentSize requires the number of bytes written to equal
	// the advertised size at Close() time. We enforce this with io.CopyN.
	c1z.ResetContentSize(outFile, dbSize)

	_, err = io.CopyN(c1z, dbFile, dbSize)
	if err != nil {
		// Always close encoder to release resources. Don't return to pool - may be in bad state.
		_ = c1z.Close()
		return fmt.Errorf("failed to copy db file into encoder: %w", err)
	}

	err = c1z.Flush()
	if err != nil {
		_ = c1z.Close()
		return fmt.Errorf("failed to flush c1z: %w", err)
	}
	err = c1z.Close()
	if err != nil {
		// Close failed, don't return to pool.
		return fmt.Errorf("failed to close c1z: %w", err)
	}

	// Successfully finished - return encoder to pool if it has pool-compatible settings.
	// This ensures the pool grows even when initially empty.
	if encoderConcurrency == pooledEncoderConcurrency {
		putEncoder(c1z)
	}

	err = outFile.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync out file: %w", err)
	}

	err = outFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close out file: %w", err)
	}
	outFile = nil

	err = dbFile.Close()
	if err != nil {
		return fmt.Errorf("failed to close db file: %w", err)
	}
	dbFile = nil

	// Atomically replace the original file with the temp file.
	// This ensures the original file remains intact if there was any
	// error during the write process.
	err = os.Rename(tmpPath, outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to rename temp file to output file: %w", err)
	}
	success = true

	// Record the decompressed and compressed sizes for every saved c1z.
	// Operators rely on this line to track c1z growth per tenant/connector
	// and to detect bloat before decoders hit the max-decoded-size cap.
	compressedSize := int64(-1)
	if outStat, statErr := os.Stat(outputFilePath); statErr == nil {
		compressedSize = outStat.Size()
	}
	zap.L().Info("c1z: saved",
		zap.String("output_path", outputFilePath),
		zap.Int64("decompressed_bytes", dbSize),
		zap.Int64("compressed_bytes", compressedSize),
	)

	return nil
}
