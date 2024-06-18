package notice

type MailSenderExt struct {
	mailSender *MailSender
	message    *Message
	tos        []string
}

func NewMailSenderExt(mailSender *MailSender, tos []string) *MailSenderExt {
	return &MailSenderExt{
		mailSender: mailSender,
		message:    NewMessage(),
		tos:        tos,
	}
}

func (m *MailSenderExt) Type() MessageType {
	return EMAIL
}

func (m *MailSenderExt) SendMessageNow(content string) (string, error) {
	for _, to := range m.tos {
		err := m.mailSender.SendEmail("backup-go", to, "备份消息通知", content)
		if err != nil {
			return "", err
		}
	}

	return "", nil
}

func (m *MailSenderExt) AddMessage(content string) {
	m.message.Add(content)
}

func (m *MailSenderExt) SendAddedMessage() (string, error) {
	content := m.message.String("<br/>")
	defer m.message.Clean()

	return m.SendMessageNow(content)
}
