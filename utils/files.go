package utils

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ZipPath(source string, target string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat source path failed: %w", err)
	}

	if !info.IsDir() {
		return "", errors.New("source path is not a directory")
	}

	baseDir := filepath.Base(source)
	zipfile, err := os.Create(target)
	if err != nil {
		return "", fmt.Errorf("create zip file failed: %w", err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk failed: %w", err)
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return fmt.Errorf("create file header failed: %w", err)
		}

		relPath, err := filepath.Rel(source, path)
		if err != nil {
			return fmt.Errorf("rel path failed: %w", err)
		}

		header.Name = filepath.Join(baseDir, relPath)
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return fmt.Errorf("create header failed: %w", err)
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open file failed: %w", err)
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("copy file failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return "", fmt.Errorf("zip failed: %w", err)
	}

	return target, nil
}
