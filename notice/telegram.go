package notice

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type TGBot struct {
	botApiFormat string
	httpClient   http.Client
}

func NewTgBot(tgKey string) TGBot {
	return TGBot{
		botApiFormat: "https://api.telegram.org/bot" + tgKey + "/%s",
		httpClient: http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
}

func (tgBot *TGBot) SendMessage(chatId, message string) (respStr string, err error) {
	url := fmt.Sprintf(tgBot.botApiFormat, "sendMessage")

	jsonData := map[string]string{"chat_id": chatId, "text": message}
	jsonValue, err := json.Marshal(jsonData)
	if err != nil {
		return
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := tgBot.httpClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	return string(body), nil
}

// Noticeable impl

type TGNotice struct {
	tg      *TGBot
	message *Message
	to      string
}

func NewTGExt(tg *TGBot, to string) *TGNotice {
	return &TGNotice{
		tg:      tg,
		message: NewMessage(),
		to:      to,
	}
}

func (m *TGNotice) Type() MessageType {
	return TG
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
