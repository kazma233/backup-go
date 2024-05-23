package main

import "testing"

func TestSendMessage(t *testing.T) {
	resp, err := SendMessage(Config.TgKey, Config.TgChatId, "test 2")
	t.Errorf("resp %v, err %v", resp, err)
}
