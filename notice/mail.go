package notice

import (
	"backup-go/utils"
	"errors"
	"log"
)

type MailNotifier struct {
	mailSender *utils.MailSender
	tos        []string
}

func NewMailNotifier(mailSender *utils.MailSender, tos []string) *MailNotifier {
	return &MailNotifier{
		mailSender: mailSender,
		tos:        tos,
	}
}

func (m *MailNotifier) IsAvailable() bool {
	return m.mailSender != nil && len(m.tos) > 0
}

func (m *MailNotifier) GetName() string {
	return "Mail"
}

func (m *MailNotifier) Send(message Message) error {
	content := message.String("<br/>")

	// 发送邮件
	errs := []error{}
	for _, to := range m.tos {
		err := m.mailSender.SendEmail("backup-go", to, "备份消息通知", content)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		log.Printf("Failed to send email: %v", errs)
		return errors.Join(errs...)
	}

	return nil
}
