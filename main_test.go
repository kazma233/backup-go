package main

import (
	"backup-go/config"
	"backup-go/utils"
	"testing"
)

func before() {
	config.InitConfig()
}

func Test_backup(t *testing.T) {
	before()

	c := defaultHolder("test", config.BackupConfig{
		BackPath: "~/Downloads/MapleMonoNormalNL-TTF",
	})

	logger := utils.NewTaskLogger(c.ID)
	c.backupWithLogger(logger)
	c.sendMessages(logger)
}

func Test_cleanOld(t *testing.T) {
	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	th.cleanHistory()
}
