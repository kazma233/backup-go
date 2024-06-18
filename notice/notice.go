package notice

import "backup-go/config"

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
			break
		case TG:
			return noticeable.SendMessageNow(content)
		}
	}

	return "", nil
}
