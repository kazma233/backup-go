package main

import (
	"backup-go/config"
	"fmt"
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

	c := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})

	c.sendMessage(fmt.Sprintf("【%s】backupTask start", c.ID))

	defer func() {
		if anyData := recover(); anyData != nil {
			c.sendMessageExt(fmt.Sprintf("【%s】backupTask has panic %v", c.ID, anyData), true)
		} else {
			c.sendMessageExt(fmt.Sprintf("【%s】backupTask finish", c.ID), true)
		}
	}()

	c.backup()
}

func Test_cleanOld(t *testing.T) {
	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	th.cleanHistory()
}
