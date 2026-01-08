package dotc1z

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"syscall"

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

	stat, err := os.Stat(c1zPath)
	if err != nil || stat.Size() == 0 {
		// TODO(kans): it would be nice to know more about the error....
		return dbFilePath, tmpDir, cleanupDir(nil)
	}

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

	outFile, err := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE|syscall.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer func() {
		if outFile != nil {
			err = outFile.Close()
			if err != nil {
				zap.L().Error("failed to close out file", zap.Error(err))
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
	c1z, err := zstd.NewWriter(outFile,
		zstd.WithEncoderConcurrency(encoderConcurrency),
	)
	if err != nil {
		return err
	}

	_, err = io.Copy(c1z, dbFile)
	if err != nil {
		return err
	}

	err = c1z.Flush()
	if err != nil {
		return fmt.Errorf("failed to flush c1z: %w", err)
	}
	err = c1z.Close()
	if err != nil {
		return fmt.Errorf("failed to close c1z: %w", err)
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

	return nil
}
