package notice

type TGExt struct {
	tg      *TGBot
	message *Message
	to      string
}

func NewTGExt(tg *TGBot, to string) *TGExt {
	return &TGExt{
		tg:      tg,
		message: NewMessage(),
		to:      to,
	}
}

func (m *TGExt) Type() MessageType {
	return TG
}

func (m *TGExt) SendMessageNow(content string) (string, error) {
	resp, err := m.tg.SendMessage(m.to, content)
	if err != nil {
		return resp, err
	}

	return "", nil
}

func (m *TGExt) AddMessage(content string) {
	m.message.Add(content)
}

func (m *TGExt) SendAddedMessage() (string, error) {
	content := m.message.String("\n")
	defer m.message.Clean()

	return m.SendMessageNow(content)
}
