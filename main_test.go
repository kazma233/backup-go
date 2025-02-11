package main

import (
	"backup-go/config"
	"fmt"
	"testing"
)

func before() {
	config.InitConfig()
}

func Test_backup(t *testing.T) {
	before()

	c := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})

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
}

func Test_cleanOld(t *testing.T) {
	before()

	th := defaultHolder("test", config.BackupConfig{
		BackPath: "E:/audio/asmr",
	})
	th.cleanHistory()
}
