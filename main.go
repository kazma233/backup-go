package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/robfig/cron/v3"
)

type (
	MessageType string
)

var (
	START MessageType = "备份开始"
	DONE  MessageType = "备份结束"
)

func main() {
	InitOSS()

	secondParser := cron.NewParser(
		cron.Second |
			cron.Minute |
			cron.Hour |
			cron.Dom |
			cron.Month |
			cron.DowOptional |
			cron.Descriptor,
	)
	c := cron.New(cron.WithParser(secondParser), cron.WithChain())

	taskId, err := c.AddFunc("0 20 0 * * ? ", backupTask)
	if err != nil {
		panic(err)
	}

	log.Printf("start task: %v, backup path: %v", taskId, Config.BackPath)

	c.Start()

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		backupTask()
	})
	log.Println(http.ListenAndServe(":7000", nil))
}

func backupTask() {
	defer func() {
		if anyData := recover(); anyData != nil {
			log.Printf("[WARN] exec backupTask has panic %v", anyData)
		}
	}()

	path := Config.BackPath
	notice(path, START)
	backup(path)
	cleanOld()
	notice(path, DONE)
}

func notice(path string, mt MessageType) {
	log.Printf("notice start %v", mt)
	mail := Config.Mail
	sender := NewMailSender(mail.Smtp, mail.Port, mail.User, mail.Password)

	hostname, _ := os.Hostname()
	content := fmt.Sprintf(`【%s】：%s目录：%s`, hostname, path, mt)
	err := sender.SendEmail("backup-go", Config.NoticeMail, "备份通知", content)
	if err != nil {
		panic(err)
	}
}

func cleanOld() {
	log.Println("cleanOld start")

	beforeDate := time.Now().AddDate(0, 0, -7)
	beforeYear, beforeMonth, beforeMonthOfDay := beforeDate.Year(), int(beforeDate.Month()), beforeDate.Day()

	var objects []oss.ObjectProperties
	token := ""
	for {
		resp, err := GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
		if err != nil {
			break
		}

		for _, object := range resp.Objects {
			sp := strings.Split(object.Key, "_")
			if len(sp) < 6 {
				continue
			}
			year, err := strconv.Atoi(sp[0])
			if err != nil {
				continue
			}
			month, err := strconv.Atoi(sp[1])
			if err != nil {
				continue
			}
			day, err := strconv.Atoi(sp[2])
			if err != nil {
				continue
			}

			if year < beforeYear {
				objects = append(objects, object)
			} else if year == beforeYear && month < beforeMonth {
				objects = append(objects, object)
			} else if year == beforeYear && month == beforeMonth && day < beforeMonthOfDay {
				objects = append(objects, object)
			}
		}
		if resp.IsTruncated {
			token = resp.NextContinuationToken
		} else {
			break
		}
	}

	if objects == nil || len(objects) <= 0 {
		return
	}

	var keys []string
	for _, k := range objects {
		keys = append(keys, k.Key)
	}
	deleteObjects, err := GetSlowClient().DeleteObjects(keys)
	if err != nil {
		panic(err)
	}
	log.Printf("delete result %v", deleteObjects)
}

func backup(path string) {
	log.Printf("start backup %s", path)
	zipFile, err := zipPath(path)
	if err != nil {
		panic(err)
	}
	log.Printf("zip path %s to %s done", path, zipFile)
	defer os.Remove(zipFile)

	objKey := filepath.Base(zipFile)
	err = Upload(objKey, zipFile)
	if err != nil {
		panic(err)
	}

	log.Printf("obj upload done %s", objKey)

	url, err := TempVisitLink(objKey)
	log.Printf("obj temp url is %s error %v", url, err)
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

	target := time.Now().Format("2006_01_02_15_04_") + baseDir + ".zip"
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
