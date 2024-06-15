package main

import (
	_ "embed"

	"gopkg.in/yaml.v2"
)

type (
	// GlobalConfig base config
	GlobalConfig struct {
		OSS        OssConfig   `yaml:"oss"`
		Mail       *MailConfig `yaml:"mail"`
		TG         *TGConfig   `yaml:"tg"`
		Cron       CornConfig  `yaml:"cron"`
		BackPath   string      `yaml:"back_path"`
		TgChatId   string      `yaml:"tg_chat_id"`
		NoticeMail string      `yaml:"notice_mail"`
		ID         string      `yaml:"id"`
	}

	OssConfig struct {
		BucketName      string `yaml:"bucket_name"`
		AccessKey       string `yaml:"access_key"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Endpoint        string `yaml:"endpoint"`
		FastEndpoint    string `yaml:"fast_endpoint"`
	}

	TGConfig struct {
		Key string `yaml:"key"`
	}

	MailConfig struct {
		Smtp     string `yaml:"smtp"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	}

	CornConfig struct {
		BackupTask string `yaml:"backup_task"`
		Liveness   string `yaml:"liveness"`
	}
)

//go:embed config.yml
var configBlob []byte

var (
	Config GlobalConfig
)

func InitConfig() {
	var config = GlobalConfig{}
	err := yaml.Unmarshal(configBlob, &config)
	if err != nil {
		panic(err)
	}

	Config = config
}
