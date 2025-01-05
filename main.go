package main

import (
	"backup-go/config"
	"backup-go/notice"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/robfig/cron/v3"
)

type TaskHolder struct {
	ID           string
	conf         config.BackupConfig
	oss          *OssClient
	c            *cron.Cron
	noticeHandle *notice.Notice
}

func defaultHolder(id string) *TaskHolder {
	return &TaskHolder{
		ID:           id,
		oss:          CreateOSSClient(),
		noticeHandle: notice.InitNotice(),
	}
}

func main() {
	config.InitConfig()

	secondParser := cron.NewParser(
		cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.DowOptional | cron.Descriptor,
	)
	c := cron.New(cron.WithParser(secondParser), cron.WithChain())

	for id, conf := range config.Config.BackupConf {
		dh := defaultHolder(id)
		dh.conf = conf
		dh.c = c

		backupTaskCron := conf.BackupTask
		if backupTaskCron == "" {
			backupTaskCron = "0 25 0 * * ?"
		}
		taskId, err := c.AddFunc(backupTaskCron, func() {
			dh.backupTask()
		})
		if err != nil {
			panic(err)
		}

		livenessCron := conf.Liveness
		if livenessCron != "" {
			_, err = c.AddFunc(livenessCron, func() {
				dh.sendMessage(fmt.Sprintf("live check report %v", time.Now()))
			})
			if err != nil {
				panic(err)
			}
		}

		dh.sendMessage(fmt.Sprintf("task %v add success", taskId))
	}

	c.Start()

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		// backupTask(ossClient)
	})
	log.Println(http.ListenAndServe(":7000", nil))
}

func (c *TaskHolder) backupTask() {
	defer func() {
		if anyData := recover(); anyData != nil {
			c.sendMessage(fmt.Sprintf("[WARN] exec backupTask has panic %v", anyData))
		}
	}()

	c.backup()
	c.cleanOld()
}

func (c *TaskHolder) cleanOld() {
	ossClient := c.oss
	c.sendMessage("cleanOld start")
	defer func() {
		c.sendMessage("cleanOld done")
	}()

	var objects []oss.ObjectProperties
	token := ""
	for {
		resp, err := ossClient.GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
		if err != nil {
			break
		}

		for _, object := range resp.Objects {
			need := NeedDeleteFile(c.ID, object.Key)
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
		c.sendMessage("no need delete")
		return
	}

	var keys []string
	for _, k := range objects {
		keys = append(keys, k.Key)
	}
	deleteObjects, err := ossClient.GetSlowClient().DeleteObjects(keys)
	c.sendMessage(fmt.Sprintf("delete result %v, err %v", deleteObjects, err))
}

func (c *TaskHolder) backup() {
	conf := c.conf
	c.sendMessage(fmt.Sprintf(`%s目录：备份开始`, conf.BackPath))

	path := conf.BackPath
	ossClient := c.oss

	if conf.BeforeCmd != "" {
		// 执行系统命令
		c.sendMessage(fmt.Sprintf("exec before command %s", conf.BeforeCmd))
		cmd := exec.Command("sh", "-c", conf.BeforeCmd)
		err := cmd.Run()
		if err != nil {
			c.sendMessage(fmt.Sprintf("exec before command %s error %v", conf.BeforeCmd, err))
			return
		}
	}

	zipFile, err := zipPath(path, c.ID)
	if err != nil {
		panic(err)
	}
	c.sendMessage(fmt.Sprintf("zip path %s to %s done", path, zipFile))
	defer os.Remove(zipFile)

	if conf.AfterCmd != "" {
		// 执行系统命令
		c.sendMessage(fmt.Sprintf("exec after command %s", conf.AfterCmd))
		cmd := exec.Command("sh", "-c", conf.AfterCmd)
		err := cmd.Run()
		if err != nil {
			c.sendMessage(fmt.Sprintf("exec after command %s error %v", conf.AfterCmd, err))
		}
	}

	objKey := filepath.Base(zipFile)
	err = ossClient.Upload(objKey, zipFile, func(message string) {
		c.sendMessage(message)
	})
	if ossClient.HasError(err) {
		panic(err)
	}

	if ossClient.HasCoolDownError(err) {
		c.sendMessage(fmt.Sprintf("obj %s upload not success, because of cool down", objKey))
	} else {
		c.sendMessage(fmt.Sprintf("obj %s upload done", objKey))
		url, err := ossClient.TempVisitLink(objKey)
		c.sendMessage(fmt.Sprintf("obj temp url: %s, error: %v", url, err))
	}

	c.sendMessageExt(fmt.Sprintf(`%s目录：备份结束`, path), true)
}

func (c *TaskHolder) sendMessage(message string) {
	c.sendMessageExt(message, false)
}

func (c *TaskHolder) sendMessageExt(message string, over bool) {
	msg := fmt.Sprintf("【备份服务(%s)通知】%s", c.ID, message)
	log.Println(msg)
	resp, err := c.noticeHandle.SendMessage(msg, over)
	if err != nil {
		log.Printf("sendNotice resp %s, error %v", resp, err)
	}
}
