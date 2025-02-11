package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func NeedDeleteFile(prefix, name string) bool {
	fp := NewProcessor()
	if err := fp.Parse(name); err != nil || !strings.EqualFold(fp.Prefix, prefix) {
		return false
	}

	fileDate := time.Date(fp.Year, time.Month(fp.Month), fp.Day, 0, 0, 0, 0, time.UTC)
	beforeDate := time.Now().AddDate(0, 0, -7)

	return fileDate.Before(beforeDate)
}

func GetFileName(prefix string) string {
	return NewProcessor().Generate(prefix, time.Now()) + ".zip"
}

func zipPath(source string, prefix string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("stat source path failed: %w", err)
	}

	if !info.IsDir() {
		return "", errors.New("source path is not a directory")
	}

	baseDir := filepath.Base(source)
	target := GetFileName(prefix)

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

// FileNameProcessor

// FileNameProcessor 结构体，用于处理字符串
type FileNameProcessor struct {
	rg     *regexp.Regexp // match string
	format string
	// parse data
	Prefix string
	Year   int
	Month  int
	Day    int
}

func NewProcessor() *FileNameProcessor {
	return &FileNameProcessor{
		rg:     regexp.MustCompile(`^([a-zA-Z0-9_]+)_(\d{4})_(\d{2})_(\d{2})`),
		format: `%s_%d_%02d_%02d`,
	}
}

// Generate 生成包含前缀和日期的字符串
func (sp *FileNameProcessor) Generate(prefix string, t time.Time) string {
	return fmt.Sprintf(sp.format, prefix, t.Year(), t.Month(), t.Day())
}

// Parse 解析包含前缀和日期的字符串，并填充结构体
func (sp *FileNameProcessor) Parse(s string) error {
	// 正则表达式匹配前缀和日期，忽略后面的任何字符
	matches := sp.rg.FindStringSubmatch(s)

	if matches == nil {
		return errors.New("invalid string format")
	}

	sp.Prefix = matches[1]
	year, err := strconv.Atoi(matches[2])
	if err != nil {
		return err
	}
	sp.Year = year

	month, err := strconv.Atoi(matches[3])
	if err != nil || month < 1 || month > 12 {
		return errors.New("invalid month value")
	}
	sp.Month = month

	day, err := strconv.Atoi(matches[4])
	if err != nil || day < 1 || day > 31 {
		return errors.New("invalid day value")
	}
	sp.Day = day

	return nil
}
