package main

import "testing"

func TestSendMessage(t *testing.T) {
	tgBot := NewTgBot(Config.TgKey)
	resp, err := tgBot.SendMessage(Config.TgChatId, "new")
	t.Errorf("resp %v, err %v", resp, err)
}
