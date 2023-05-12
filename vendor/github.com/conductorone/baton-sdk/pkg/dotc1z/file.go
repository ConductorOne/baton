package dotc1z

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"github.com/klauspost/compress/zstd"
)

func loadC1z(filePath string) (string, error) {
	workingDir, err := os.MkdirTemp("", "c1z")
	if err != nil {
		return "", err
	}
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

		r, err := NewDecoder(c1zFile)
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

func saveC1z(dbFilePath string, outputFilePath string) error {
	if outputFilePath == "" {
		return errors.New("c1z: output file path not configured")
	}

	dbFile, err := os.Open(dbFilePath)
	if err != nil {
		return err
	}
	defer dbFile.Close()

	outFile, err := os.OpenFile(outputFilePath, os.O_RDWR|os.O_CREATE|syscall.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Write the magic file header
	_, err = outFile.Write(C1ZFileHeader)
	if err != nil {
		return err
	}

	c1z, err := zstd.NewWriter(outFile)
	if err != nil {
		return err
	}

	_, err = io.Copy(c1z, dbFile)
	if err != nil {
		return err
	}

	err = c1z.Flush()
	if err != nil {
		return err
	}
	err = c1z.Close()
	if err != nil {
		return err
	}

	return nil
}
