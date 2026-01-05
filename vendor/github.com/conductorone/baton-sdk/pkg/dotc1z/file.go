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

func loadC1z(filePath string, tmpDir string, opts ...DecoderOption) (string, error) {
	var err error
	workingDir, err := os.MkdirTemp(tmpDir, "c1z")
	if err != nil {
		return "", err
	}
	defer func() {
		if err != nil {
			if removeErr := os.RemoveAll(workingDir); removeErr != nil {
				err = errors.Join(err, removeErr)
			}
		}
	}()
	dbFilePath := filepath.Join(workingDir, "db")
	dbFile, err := os.Create(dbFilePath)
	if err != nil {
		return "", err
	}
	defer dbFile.Close()

	if stat, err := os.Stat(filePath); err == nil && stat.Size() != 0 {
		c1zFile, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer c1zFile.Close()

		r, err := NewDecoder(c1zFile, opts...)
		if err != nil {
			return "", err
		}
		_, err = io.Copy(dbFile, r)
		if err != nil {
			return "", err
		}
		err = r.Close()
		if err != nil {
			return "", err
		}
	}

	return dbFilePath, nil
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
