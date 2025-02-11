package notice

import "backup-go/utils"

type MailNotice struct {
	mailSender *utils.MailSender
	message    *Message
	tos        []string
}

func NewMailSenderExt(mailSender *utils.MailSender, tos []string) *MailNotice {
	return &MailNotice{
		mailSender: mailSender,
		message:    NewMessage(),
		tos:        tos,
	}
}

func (m *MailNotice) SendMessageNow(content string) (string, error) {
	for _, to := range m.tos {
		err := m.mailSender.SendEmail("backup-go", to, "备份消息通知", content)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func (m *MailNotice) AddMessage(content string) {
	m.message.Add(content)
}

func (m *MailNotice) SendAddedMessage() (string, error) {
	content := m.message.String("<br/>")
	defer m.message.Clean()

	return m.SendMessageNow(content)
}
