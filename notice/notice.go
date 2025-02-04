package notice

import (
	"backup-go/config"
	"fmt"
	"strings"
	"time"
)

type MessageType int

const (
	EMAIL MessageType = 1
	TG    MessageType = 2
)

type Noticeable interface {
	Type() MessageType
	SendMessageNow(content string) (string, error)
	AddMessage(content string)
	SendAddedMessage() (string, error)
}

type Notice struct {
	noticeHandle []Noticeable
}

func InitNotice() *Notice {
	var noticeHandle []Noticeable

	if config.Config.TG != nil {
		tb := NewTgBot(config.Config.TG.Key)
		noticeHandle = append(noticeHandle, NewTGExt(&tb, config.Config.TgChatId))
	}

	mailConfig := config.Config.Mail
	if mailConfig != nil {
		ms := NewMailSender(mailConfig.Smtp, mailConfig.Port, mailConfig.User, mailConfig.Password)
		noticeHandle = append(noticeHandle, NewMailSenderExt(&ms, config.Config.NoticeMail))
	}

	return &Notice{noticeHandle: noticeHandle}
}

func (n *Notice) SendMessage(content string, over bool) (string, error) {
	for _, noticeable := range n.noticeHandle {
		switch noticeable.Type() {
		case EMAIL:
			if over {
				noticeable.AddMessage(content)
				return noticeable.SendAddedMessage()
			}
			noticeable.AddMessage(content)
		case TG:
			return noticeable.SendMessageNow(content)
		}
	}

	return "", nil
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
