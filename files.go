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

func NeedDeleteFile(name string) bool {
	fp := FileNameProcessor{}
	err := fp.Parse(name)
	if err != nil {
		return false
	}

	year, month, day := fp.year, fp.month, fp.day
	beforeDate := time.Now().AddDate(0, 0, -7)
	beforeYear, beforeMonth, beforeMonthOfDay := beforeDate.Year(), int(beforeDate.Month()), beforeDate.Day()

	if year < beforeYear {
		return true
	} else if year == beforeYear && month < beforeMonth {
		return true
	} else if year == beforeYear && month == beforeMonth && day < beforeMonthOfDay {
		return true
	}

	return false
}

func GetFileName() string {
	return FileNameProcessor{}.Generate(ID, time.Now()) + ".zip"
}

// FileNameProcessor 结构体，用于处理字符串
type FileNameProcessor struct {
	prefix string
	year   int
	month  int
	day    int
	rg     *regexp.Regexp // match string
}

// Generate 生成包含前缀和日期的字符串
func (sp FileNameProcessor) Generate(prefix string, t time.Time) string {
	return fmt.Sprintf("%s_%d_%02d_%02d", prefix, t.Year(), t.Month(), t.Day())
}

// Parse 解析包含前缀和日期的字符串，并填充结构体
func (sp FileNameProcessor) Parse(s string) error {
	// 正则表达式匹配前缀和日期，忽略后面的任何字符
	re := regexp.MustCompile(`^([^_]+)_(\d{4})(\d{2})(\d{2})`)
	matches := re.FindStringSubmatch(s)

	if matches == nil {
		return errors.New("invalid string format")
	}

	sp.prefix = matches[1]
	year, err := strconv.Atoi(matches[2])
	if err != nil {
		return err
	}
	sp.year = year

	month, err := strconv.Atoi(matches[3])
	if err != nil || month < 1 || month > 12 {
		return errors.New("invalid month value")
	}
	sp.month = month

	day, err := strconv.Atoi(matches[4])
	if err != nil || day < 1 || day > 31 {
		return errors.New("invalid day value")
	}
	sp.day = day

	return nil
}

func zipPath(source string) (string, error) {
	info, err := os.Stat(source)
	if err != nil {
		panic(err)
	}

	if !info.IsDir() {
		panic(errors.New("path is not dir"))
	}
	baseDir := filepath.Base(source)

	target := GetFileName()
	zipfile, err := os.Create(target)
	if err != nil {
		panic(err)
	}
	defer zipfile.Close()

	archive := zip.NewWriter(zipfile)
	defer archive.Close()

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		n := baseDir + filepath.ToSlash(strings.TrimPrefix(path, source))
		if n == "" {
			return nil
		}
		header.Name = n

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}

		writer, err := archive.CreateHeader(header)
		if err != nil {
			return err
		}

		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			_, err = io.Copy(writer, file)
			return err
		}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return target, nil
}
