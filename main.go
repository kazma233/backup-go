package main

import (
	"archive/zip"
	"errors"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/robfig/cron/v3"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
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

	c.Run()
}

func backupTask() {
	defer func() {
		if anyData := recover(); anyData != nil {
			log.Printf("[WARN] exec backupTask has panic %v", anyData)
		}
	}()

	path := Config.BackPath
	backup(path)
	cleanOld()
	notice(path)
}

func notice(path string) {
	mail := Config.Mail
	sender := NewMailSender(mail.Smtp, mail.Port, mail.User, mail.Password)
	err := sender.SendEmail(mail.User, Config.NoticeMail, "备份通知", path+"已备份完成")
	if err != nil {
		panic(err)
	}
}

func cleanOld() {
	client, err := oss.New(Config.OSS.Endpoint, Config.OSS.AccessKey, Config.OSS.AccessKeySecret)
	if err != nil {
		panic(err)
	}

	bucket, err := client.Bucket(Config.OSS.BucketName)
	if err != nil {
		panic(err)
	}

	objects, err := bucket.ListObjectsV2(oss.Prefix(time.Now().AddDate(0, 0, -7).Format("2006_01_02")))
	if err != nil {
		panic(err)
	}

	if objects.Objects == nil || len(objects.Objects) < 0 {
		return
	}

	var keys []string
	for _, k := range objects.Objects {
		keys = append(keys, k.Key)
	}
	deleteObjects, err := bucket.DeleteObjects(keys)
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

	client, err := oss.New(Config.OSS.Endpoint, Config.OSS.AccessKey, Config.OSS.AccessKeySecret)
	if err != nil {
		panic(err)
	}

	bucket, err := client.Bucket(Config.OSS.BucketName)
	if err != nil {
		panic(err)
	}

	objKey := filepath.Base(zipFile)
	err = bucket.PutObjectFromFile(objKey, zipFile)
	if err != nil {
		panic(err)
	}

	log.Printf("obj upload done %s", objKey)

	url, err := bucket.SignURL(objKey, oss.HTTPGet, 60*60*24*7)
	if err != nil {
		panic(err)
	}

	log.Printf("obj temp url is %s", url)
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

	target := time.Now().Format("2006_01_02_15_04") + baseDir + ".zip"
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
