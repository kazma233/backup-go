package notice

import (
	"log"
)

type Notifier interface {
	// Send 发送消息
	Send(msg Message) error

	// IsAvailable 检查通知渠道是否可用
	IsAvailable() bool

	// GetName 获取通知渠道名称
	GetName() string
}

type NoticeManager struct {
	notifiers []Notifier
	msgBuffer Message
}

func NewNoticeManager() *NoticeManager {
	return &NoticeManager{
		notifiers: make([]Notifier, 0),
		msgBuffer: Message{},
	}
}

func (m *NoticeManager) AddNotifier(n Notifier) {
	m.notifiers = append(m.notifiers, n)
}

func (m *NoticeManager) AddMessage2Queue(msg string) {
	m.msgBuffer.Add(msg)
}

func (m *NoticeManager) Notice() {
	for _, n := range m.notifiers {
		if n.IsAvailable() {
			continue
		}

		if err := n.Send(m.msgBuffer); err != nil {
			log.Printf("Failed to send messages via %s: %v", n.GetName(), err)
		}
	}

	m.msgBuffer.Clean()
}
