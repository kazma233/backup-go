package main

import (
	"backup-go/config"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/robfig/cron/v3"
)

func main() {
	config.InitConfig()
	InitNotice()

	secondParser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor,
	)
	c := cron.New(cron.WithParser(secondParser), cron.WithChain())

	ossClient := CreateOSSClient()

	backupTaskCron := config.Config.Cron.BackupTask
	if backupTaskCron == "" {
		backupTaskCron = "0 25 0 * * ?"
	}
	taskId, err := c.AddFunc(backupTaskCron, func() {
		backupTask(ossClient)
	})
	if err != nil {
		panic(err)
	}

	livenessCron := config.Config.Cron.Liveness
	if livenessCron != "" {
		_, err = c.AddFunc(livenessCron, func() {
			sendMessage(fmt.Sprintf("live check report %v", time.Now()))
		})
		if err != nil {
			panic(err)
		}
	}

	sendMessage(
		fmt.Sprintf("start task: %d, id %s, backup path: %s",
			taskId, config.Config.ID, config.Config.BackPath),
	)

	c.Start()

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		backupTask(ossClient)
	})
	log.Println(http.ListenAndServe(":7000", nil))
}

func backupTask(ossClient *OssClient) {
	defer func() {
		if anyData := recover(); anyData != nil {
			sendMessage(fmt.Sprintf("[WARN] exec backupTask has panic %v", anyData))
		}
	}()

	path := config.Config.BackPath
	sendMessage(fmt.Sprintf(`%s目录：备份开始`, path))
	backup(path, ossClient)
	cleanOld(ossClient)
	sendMessageExt(fmt.Sprintf(`%s目录：备份结束`, path), true)
}

func cleanOld(ossClient *OssClient) {
	sendMessage("cleanOld start")
	defer func() {
		sendMessage("cleanOld done")
	}()

	var objects []oss.ObjectProperties
	token := ""
	for {
		resp, err := ossClient.GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
		if err != nil {
			break
		}

		for _, object := range resp.Objects {
			need := NeedDeleteFile(config.Config.ID, object.Key)
			if need {
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
		sendMessage("no need delete")
		return
	}

	var keys []string
	for _, k := range objects {
		keys = append(keys, k.Key)
	}
	deleteObjects, err := ossClient.GetSlowClient().DeleteObjects(keys)
	sendMessage(fmt.Sprintf("delete result %v, err %v", deleteObjects, err))
}

func backup(path string, ossClient *OssClient) {
	sendMessage(fmt.Sprintf("%s backup start", path))

	zipFile, err := zipPath(path, config.Config.ID)
	if err != nil {
		panic(err)
	}

	sendMessage(fmt.Sprintf("zip path %s to %s done", path, zipFile))
	defer os.Remove(zipFile)

	objKey := filepath.Base(zipFile)
	err = ossClient.Upload(objKey, zipFile, func(message string) {
		sendMessage(message)
	})
	if ossClient.HasError(err) {
		panic(err)
	}

	if ossClient.HasCoolDownError(err) {
		sendMessage(fmt.Sprintf("obj %s upload not success, because of cool down", objKey))
	} else {
		sendMessage(fmt.Sprintf("obj %s upload done", objKey))
		url, err := ossClient.TempVisitLink(objKey)
		sendMessage(fmt.Sprintf("obj temp url: %s, error: %v", url, err))
	}
}
