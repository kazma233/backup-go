package main

import (
	"backup-go/config"
	"log"
	"testing"
)

func before() {
	config.InitConfig()
}

func Test_zipPath(t *testing.T) {
	before()

	path, err := zipPath(`/Users/fanggeek/Projects/cayenne`, "test")
	if err != nil {
		panic(err)
	}

	log.Printf("path %s", path)
}

func Test_backup(t *testing.T) {
	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	th.backup()
}

func Test_cleanOld(t *testing.T) {
	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	th.cleanOld()
}
