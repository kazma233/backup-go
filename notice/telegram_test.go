package notice

import (
	"backup-go/config"
	"testing"
)

func TestSendMessage(t *testing.T) {
	tgBot := NewTgBot(config.Config.TG.Key)
	resp, err := tgBot.SendMessage(config.Config.TgChatId, "new")
	t.Errorf("resp %v, err %v", resp, err)
}
