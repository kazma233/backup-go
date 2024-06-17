package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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

	tgBot      *TGBot
	mailSender *MailSender

	msg = NewMessage()
)

func main() {
	InitConfig()
	InitID()

	if Config.TG != nil {
		tb := NewTgBot(Config.TG.Key)
		tgBot = &tb
	}

	mailConfig := Config.Mail
	if mailConfig != nil {
		ms := NewMailSender(mailConfig.Smtp, mailConfig.Port, mailConfig.User, mailConfig.Password)
		mailSender = &ms
	}

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

	ossClient := CreateOSSClient()

	backupTaskCron := Config.Cron.BackupTask
	if backupTaskCron == "" {
		backupTaskCron = "0 25 0 * * ?"
	}
	taskId, err := c.AddFunc(backupTaskCron, func() {
		backupTask(ossClient)
	})
	if err != nil {
		panic(err)
	}

	livenessCron := Config.Cron.Liveness
	if livenessCron != "" {
		_, err = c.AddFunc(livenessCron, func() {
			sendMessage(fmt.Sprintf("live check report %v", time.Now()))
		})
		if err != nil {
			panic(err)
		}
	}

	sendMessage(fmt.Sprintf("start task: %v, backup path: %v", taskId, Config.BackPath))

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

	path := Config.BackPath
	notice(path, START)
	backup(path, ossClient)
	cleanOld(ossClient)
	notice(path, DONE)
	mailNoticeAndClean()
}

func mailNoticeAndClean() {
	if mailSender != nil {
		mailContent := msg.String()
		defer msg.Clean()

		mailList := Config.NoticeMail
		for _, mail := range mailList {
			err := mailSender.SendEmail("backup-go", mail, "备份消息通知", mailContent)
			if err != nil {
				log.Printf("mail notice error %v", err)
			}
		}
	}
}

func notice(path string, mt MessageType) {
	message := fmt.Sprintf(`备份通知：【%s】：%s目录：%s`, ID, path, mt)
	sendMessage(message)
}

func sendMessage(message string) {
	log.Println(message)
	if tgBot != nil {
		resp, err := tgBot.SendMessage(Config.TgChatId, message)
		log.Printf("tg notice resp %s error %v", resp, err)
	}
	msg.Add(message)
}

func cleanOld(ossClient *OssClient) {
	sendMessage("cleanOld start")

	var objects []oss.ObjectProperties
	token := ""
	for {
		resp, err := ossClient.GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
		if err != nil {
			break
		}

		for _, object := range resp.Objects {
			need := NeedDeleteFile(object.Key)
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
		return
	}

	var keys []string
	for _, k := range objects {
		keys = append(keys, k.Key)
	}
	deleteObjects, err := ossClient.GetSlowClient().DeleteObjects(keys)
	if err != nil {
		panic(err)
	}

	sendMessage(fmt.Sprintf("delete result %v", deleteObjects))
}

func backup(path string, ossClient *OssClient) {
	sendMessage(fmt.Sprintf("start backup %s", path))

	zipFile, err := zipPath(path)
	if err != nil {
		panic(err)
	}

	sendMessage(fmt.Sprintf("zip path %s to %s done", path, zipFile))
	defer os.Remove(zipFile)

	objKey := filepath.Base(zipFile)
	err = ossClient.Upload(objKey, zipFile, func(message string) {
		sendMessage(message)
	})
	if err != nil {
		panic(err)
	}

	sendMessage(fmt.Sprintf("obj upload done %s", objKey))

	url, err := ossClient.TempVisitLink(objKey)
	sendMessage(fmt.Sprintf("obj temp url is %s error %v", url, err))
}
