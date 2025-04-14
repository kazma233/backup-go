package main

import (
	"backup-go/config"
	"backup-go/notice"
	"backup-go/utils"
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
	ID            string
	conf          config.BackupConfig
	oss           *OssClient
	noticeManager *notice.NoticeManager
}

func defaultHolder(id string, conf config.BackupConfig) *TaskHolder {
	if id == "" || conf.BackPath == "" {
		panic("id or back_path can not be empty")
	}

	nm := notice.NewNoticeManager()
	if config.Config.TG != nil {
		tgBot := utils.NewTgBot(config.Config.TG.Key)
		nm.AddNotifier(notice.NewTGNotifier(&tgBot, config.Config.TgChatId))
	}
	if config.Config.Mail != nil {
		mailConfig := config.Config.Mail
		ms := utils.NewMailSender(mailConfig.Smtp, mailConfig.Port, mailConfig.User, mailConfig.Password)
		nm.AddNotifier(notice.NewMailNotifier(&ms, config.Config.NoticeMail))
	}

	return &TaskHolder{
		ID:            id,
		conf:          conf,
		oss:           CreateOSSClient(config.Config.OSS),
		noticeManager: nm,
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

		log.Printf("task %v add success", taskId)
	}

	c.Start()

	http.HandleFunc("/", func(resp http.ResponseWriter, req *http.Request) {
		id := req.URL.Query().Get("id")
		dh := defaultHolder(id, config.Config.BackupConf[id])
		log.Printf("backup task %v", dh)

		dh.backupTask()
	})
	log.Println(http.ListenAndServe(":7000", nil))
}

func (c *TaskHolder) backupTask() {
	c.addMessage(fmt.Sprintf("【%s】backupTask start", c.ID))

	defer func() {
		if anyData := recover(); anyData != nil {
			c.addMessage(fmt.Sprintf("【%s】backupTask has panic %v", c.ID, anyData))
		} else {
			c.addMessage(fmt.Sprintf("【%s】backupTask finish", c.ID))
		}
		c.sendMessage()
	}()

	c.backup()
	c.cleanHistory()
}

func (c *TaskHolder) cleanHistory() {
	ossClient := c.oss
	c.addMessage("clean history start")
	defer func() {
		c.addMessage("clean history done")
	}()

	var objects []oss.ObjectProperties
	token := ""
	for {
		resp, err := ossClient.GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
		if err != nil {
			break
		}

		for _, object := range resp.Objects {
			need := IsNeedDeleteFile(c.ID, object.Key)
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

	if len(objects) <= 0 {
		c.addMessage("no need delete")
		return
	}

	var keys []string
	for _, k := range objects {
		keys = append(keys, k.Key)
	}
	deleteObjects, err := ossClient.GetSlowClient().DeleteObjects(keys)
	if err != nil {
		c.addMessage(fmt.Sprintf("delete has err: %v", err))
	} else {
		c.addMessage(fmt.Sprintf("delete success, deleteObjects is %v", deleteObjects))
	}
}

func (c *TaskHolder) backup() {
	conf := c.conf
	path := conf.BackPath

	c.addMessage(fmt.Sprintf(`backup【%s】start`, path))
	defer func() {
		c.addMessage(fmt.Sprintf(`backup【%s】done`, path))
	}()

	if conf.BeforeCmd != "" {
		c.addMessage(fmt.Sprintf("exec before command: 【%s】", conf.BeforeCmd))
		cmd := exec.Command("sh", "-c", conf.BeforeCmd)
		err := cmd.Run()
		if err != nil {
			c.addMessage(fmt.Sprintf("exec before command【%s】has error【%s】", conf.BeforeCmd, err))
			return
		}
	}

	zipFile, err := utils.ZipPath(path, GetFileName(c.ID), func(filePath string, processed, total int64, percentage float64) {
		// c.addMessage(fmt.Sprintf("zip %s: %d/%d (%.2f%%)", filePath, processed, total, percentage))
		log.Printf("zip %s: %d/%d (%.2f%%)", filePath, processed, total, percentage)
	})
	if err != nil {
		panic(err)
	}
	c.addMessage(fmt.Sprintf("zip path【%s】to【%s】done", path, zipFile))
	defer os.Remove(zipFile)

	if conf.AfterCmd != "" {
		c.addMessage(fmt.Sprintf("exec after command: 【%s】", conf.AfterCmd))
		cmd := exec.Command("sh", "-c", conf.AfterCmd)
		err := cmd.Run()
		if err != nil {
			c.addMessage(fmt.Sprintf("exec after command【%s】has error【%s】", conf.AfterCmd, err))
			return
		}
	}

	objKey := filepath.Base(zipFile)
	ossClient := c.oss
	err = ossClient.Upload(objKey, zipFile, func(message string) {
		c.addMessage(message)
	})
	if ossClient.HasError(err) {
		panic(err)
	}

	if ossClient.HasCoolDownError(err) {
		c.addMessage(fmt.Sprintf("obj【%s】upload not success, because of cool down", objKey))
	} else {
		c.addMessage(fmt.Sprintf("obj【%s】upload done", objKey))
	}
}

func (c *TaskHolder) addMessage(message string) {
	log.Println(message)
	c.noticeManager.AddMessage2Queue(message)
}

func (c *TaskHolder) sendMessage() {
	c.noticeManager.Notice()
}
