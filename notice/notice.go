package notice

import (
	"backup-go/config"
	"backup-go/utils"
	"fmt"
	"strings"
	"time"
)

type Noticeable interface {
	AddMessage(content string)
	SendAddedMessage() (string, error)
}

type Notice struct {
	noticeHandle []Noticeable
}

func InitNotice() *Notice {
	var noticeHandle []Noticeable

	if config.Config.TG != nil {
		tb := utils.NewTgBot(config.Config.TG.Key)
		noticeHandle = append(noticeHandle, NewTGExt(&tb, config.Config.TgChatId))
	}

	mailConfig := config.Config.Mail
	if mailConfig != nil {
		ms := utils.NewMailSender(mailConfig.Smtp, mailConfig.Port, mailConfig.User, mailConfig.Password)
		noticeHandle = append(noticeHandle, NewMailSenderExt(&ms, config.Config.NoticeMail))
	}

	return &Notice{noticeHandle: noticeHandle}
}

func (n *Notice) AddMessage(content string, over bool) {
	for _, noticeable := range n.noticeHandle {
		noticeable.AddMessage(content)
	}
}

func (n *Notice) SendMessage() {
	for _, noticeable := range n.noticeHandle {
		resp, err := noticeable.SendAddedMessage()
		if err != nil {
			fmt.Printf("sendNotice resp %s, error %v", resp, err)
		}
	}
}

// Message

type MessageBody struct {
	Content string
	Date    time.Time
}

type Message struct {
	messageItems []MessageBody
}

func NewMessage() *Message {
	return &Message{
		messageItems: make([]MessageBody, 0),
	}
}

func (m *Message) String(sep string) string {
	if m.messageItems == nil || len(m.messageItems) <= 0 {
		return ""
	}

	result := []string{}
	for _, m := range m.messageItems {
		result = append(result, fmt.Sprintf("%s: %s", m.Date.Format("2006-01-02 15:04:05"), m.Content))
	}

	return strings.Join(result, sep)
}

func (m *Message) Add(content string) {
	m.messageItems = append(m.messageItems, MessageBody{
		Content: content,
		Date:    time.Now().Local(),
	})
}

func (m *Message) Clean() {
	m.messageItems = make([]MessageBody, 0)
}
