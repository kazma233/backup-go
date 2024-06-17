package main

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const FIX_STEP = "_"
const DATE_FORMAT = "2006_01_02_15_04_05"

func NeedDeleteFile(name string) bool {
	sp := strings.Split(name, FIX_STEP)
	if len(sp) < 7 {
		return false
	}

	dateStartIndex := 1
	year, err := strconv.Atoi(sp[dateStartIndex])
	if err != nil {
		return false
	}

	dateStartIndex += 1
	month, err := strconv.Atoi(sp[dateStartIndex])
	if err != nil {
		return false
	}
	dateStartIndex += 1
	day, err := strconv.Atoi(sp[dateStartIndex])
	if err != nil {
		return false
	}

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
	return ID + FIX_STEP + time.Now().Format(DATE_FORMAT) + ".zip"
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
