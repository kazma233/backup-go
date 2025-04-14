package notice

import (
	"fmt"
	"strings"
	"time"
)

type MessageBody struct {
	Content string
	Date    time.Time
}

type Message struct {
	messageItems []MessageBody
}

func (m *Message) String(sep string) string {
	if len(m.messageItems) <= 0 {
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
		Date:    time.Now(),
	})
}

func (m *Message) Clean() {
	m.messageItems = make([]MessageBody, 0)
}
