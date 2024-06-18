package main

import (
	"backup-go/config"
	"backup-go/notice"
	"fmt"
	"log"
)

var (
	noticeHandle *notice.Notice
)

func InitNotice() {
	noticeHandle = notice.InitNotice()
}

func sendMessage(message string) {
	sendMessageExt(message, false)
}

func sendMessageExt(message string, over bool) {
	msg := fmt.Sprintf("【备份服务(%s)通知】%s", config.Config.ID, message)
	log.Println(msg)
	resp, err := noticeHandle.SendMessage(msg, over)
	if err != nil {
		log.Printf("sendNotice resp %s, error %v", resp, err)
	}
}
