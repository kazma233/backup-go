package notice

import "backup-go/utils"

type TGNotice struct {
	tg      *utils.TGBot
	message *Message
	to      string
}

func NewTGExt(tg *utils.TGBot, to string) *TGNotice {
	return &TGNotice{
		tg:      tg,
		message: NewMessage(),
		to:      to,
	}
}

func (m *TGNotice) SendMessageNow(content string) (string, error) {
	resp, err := m.tg.SendMessage(m.to, content)
	if err != nil {
		return resp, err
	}

	return "", nil
}

func (m *TGNotice) AddMessage(content string) {
	m.message.Add(content)
}

func (m *TGNotice) SendAddedMessage() (string, error) {
	content := m.message.String("\n")
	defer m.message.Clean()

	return m.SendMessageNow(content)
}
