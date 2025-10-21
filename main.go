package main

import (
	"backup-go/config"
	"backup-go/notice"
	"backup-go/utils"
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
	logger := utils.NewTaskLogger(c.ID)

	// 使用 TaskLogger 的装饰器方法
	logger.ExecuteWithPanic("backupTask", func() {
		logger.Execute("backup", func() error {
			c.backupWithLogger(logger)
			return nil
		})

		logger.Execute("cleanHistory", func() error {
			c.cleanHistoryWithLogger(logger)
			return nil
		})
	})

	// 在 main.go 中处理消息发送
	c.sendMessages(logger)
}

func (c *TaskHolder) cleanHistory() {
	c.cleanHistoryWithLogger(utils.NewTaskLogger(c.ID))
}

func (c *TaskHolder) cleanHistoryWithLogger(logger *utils.TaskLogger) {
	ossClient := c.oss

	var objects []oss.ObjectProperties
	token := ""
	for {
		resp, err := ossClient.GetSlowClient().ListObjectsV2(oss.MaxKeys(100), oss.ContinuationToken(token))
		if err != nil {
			logger.Log("列出对象失败: %v", err)
			break
		}

		for _, object := range resp.Objects {
			need := utils.IsNeedDeleteFile(c.ID, object.Key)
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
		logger.Log("无需删除文件")
		return
	}

	var keys []string
	for _, k := range objects {
		keys = append(keys, k.Key)
	}

	logger.Log("找到 %d 个文件需要删除", len(keys))
	deleteObjects, err := ossClient.GetSlowClient().DeleteObjects(keys)
	if err != nil {
		logger.Log("删除失败: %v", err)
	} else {
		logger.Log("成功删除：%v", deleteObjects.DeletedObjects)
	}
}

func (c *TaskHolder) backupWithLogger(logger *utils.TaskLogger) {
	conf := c.conf
	path := conf.BackPath

	logger.Log("开始备份, 备份path: %s", path)

	// 执行前置命令
	if conf.BeforeCmd != "" {
		logger.Log("执行前置命令: %s", conf.BeforeCmd)
		cmd := exec.Command("sh", "-c", conf.BeforeCmd)
		if err := cmd.Run(); err != nil {
			logger.Log("前置命令执行失败: %v", err)
			return
		}
		logger.Log("前置命令执行完成")
	}

	// 压缩文件
	zipFile, err := utils.ZipPath(path, utils.GetFileName(c.ID), func(filePath string, processed, total int64, percentage float64) {
		logger.Log("zip %s: %d/%d (%.2f%%)", filePath, processed, total, percentage)
	}, func(total int64) {
		logger.Log("压缩完成，总大小: %d 字节", total)
	})
	if err != nil {
		logger.Log("压缩失败: %v", err)
		panic(err)
	}
	defer os.Remove(zipFile)

	// 执行后置命令
	if conf.AfterCmd != "" {
		logger.Log("执行后置命令: %s", conf.AfterCmd)
		cmd := exec.Command("sh", "-c", conf.AfterCmd)
		if err := cmd.Run(); err != nil {
			logger.Log("后置命令执行失败: %v", err)
			return
		}
		logger.Log("后置命令执行完成")
	}

	// 上传到OSS
	objKey := filepath.Base(zipFile)
	ossClient := c.oss
	logger.Log("上传到OSS: %s", objKey)

	err = ossClient.Upload(objKey, zipFile, func(message string) {
		logger.Log("上传进度: %s", message)
	})

	if ossClient.HasError(err) {
		logger.Log("上传失败: %v", err)
		panic(err)
	}

	if ossClient.HasCoolDownError(err) {
		logger.Log("上传因冷却期延迟: %s", objKey)
	} else {
		logger.Log("上传完成: %s", objKey)
	}
}

// sendMessages 发送 TaskLogger 收集的所有消息
func (c *TaskHolder) sendMessages(logger *utils.TaskLogger) {
	message := logger.GetMessageAndClean()
	c.noticeManager.Notice(message)
}
