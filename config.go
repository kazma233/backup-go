package main

import (
	_ "embed"

	"gopkg.in/yaml.v2"
)

type (
	// GlobalConfig base config
	GlobalConfig struct {
		OSS      OssConfig `yaml:"oss"`
		BackPath string    `yaml:"back_path"`
	}

	OssConfig struct {
		BucketName      string `yaml:"bucket_name"`
		AccessKey       string `yaml:"access_key"`
		AccessKeySecret string `yaml:"access_key_secret"`
		Endpoint        string `yaml:"endpoint"`
	}
)

//go:embed config.yml
var configBlob []byte

var (
	Config GlobalConfig
)

func init() {
	var config = GlobalConfig{}
	err := yaml.Unmarshal(configBlob, &config)
	if err != nil {
		panic(err)
	}

	Config = config
}
