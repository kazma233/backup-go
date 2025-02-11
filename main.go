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

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/robfig/cron/v3"
)

type TaskHolder struct {
	ID           string
	conf         config.BackupConfig
	oss          *OssClient
	noticeHandle *notice.Notice
}

func defaultHolder(id string, conf config.BackupConfig) *TaskHolder {
	if id == "" || conf.BackPath == "" {
		panic("id or back_path can not be empty")
	}

	return &TaskHolder{
		ID:           id,
		conf:         conf,
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
		dh := defaultHolder(id, conf)

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

		fmt.Printf("task %v add success", taskId)
	}

	c.Start()

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		dh := defaultHolder(id, config.Config.BackupConf[id])
		fmt.Printf("backup task %v", dh)

		dh.backupTask()
	})
	log.Println(http.ListenAndServe(":7000", nil))
}

func (c *TaskHolder) backupTask() {
	c.sendMessage(fmt.Sprintf("【%s】backupTask start", c.ID))

	defer func() {
		if anyData := recover(); anyData != nil {
			c.sendMessageExt(fmt.Sprintf("【%s】backupTask has panic %v", c.ID, anyData), true)
		} else {
			c.sendMessageExt(fmt.Sprintf("【%s】backupTask finish", c.ID), true)
		}
	}()

	c.backup()
	c.cleanHistory()
}

func (c *TaskHolder) cleanHistory() {
	ossClient := c.oss
	c.sendMessage("clean history start")
	defer func() {
		c.sendMessage("clean history done")
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
	if err != nil {
		c.sendMessage(fmt.Sprintf("delete has err: %v", err))
	} else {
		c.sendMessage(fmt.Sprintf("delete success, deleteObjects is %v", deleteObjects))
	}
}

func (c *TaskHolder) backup() {
	conf := c.conf
	path := conf.BackPath

	c.sendMessage(fmt.Sprintf(`backup【%s】start`, path))
	defer func() {
		c.sendMessage(fmt.Sprintf(`backup【%s】done`, path))
	}()

	if conf.BeforeCmd != "" {
		c.sendMessage(fmt.Sprintf("exec before command: 【%s】", conf.BeforeCmd))
		cmd := exec.Command("sh", "-c", conf.BeforeCmd)
		err := cmd.Run()
		if err != nil {
			c.sendMessage(fmt.Sprintf("exec before command【%s】has error【%s】", conf.BeforeCmd, err))
			return
		}
	}

	zipFile, err := zipPath(path, c.ID)
	if err != nil {
		panic(err)
	}
	c.sendMessage(fmt.Sprintf("zip path【%s】to【%s】done", path, zipFile))
	defer os.Remove(zipFile)

	if conf.AfterCmd != "" {
		c.sendMessage(fmt.Sprintf("exec after command: 【%s】", conf.AfterCmd))
		cmd := exec.Command("sh", "-c", conf.AfterCmd)
		err := cmd.Run()
		if err != nil {
			c.sendMessage(fmt.Sprintf("exec after command【%s】has error【%s】", conf.AfterCmd, err))
			return
		}
	}

	objKey := filepath.Base(zipFile)
	ossClient := c.oss
	err = ossClient.Upload(objKey, zipFile, func(message string) {
		c.sendMessage(message)
	})
	if ossClient.HasError(err) {
		panic(err)
	}

	if ossClient.HasCoolDownError(err) {
		c.sendMessage(fmt.Sprintf("obj【%s】upload not success, because of cool down", objKey))
	} else {
		c.sendMessage(fmt.Sprintf("obj【%s】upload done", objKey))
	}
}

func (c *TaskHolder) sendMessage(message string) {
	c.sendMessageExt(message, false)
}

func (c *TaskHolder) sendMessageExt(message string, over bool) {
	log.Println(message)
	resp, err := c.noticeHandle.SendMessage(message, over)
	if err != nil {
		log.Printf("sendNotice resp %s, error %v", resp, err)
	}
}
