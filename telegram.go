package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	botFormat  = "https://api.telegram.org/bot%s/%s"
	httpClient = http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
)

func SendMessage(tgKey, chatId, message string) (respStr string, err error) {
	url := fmt.Sprintf(botFormat, tgKey, "sendMessage")

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

	resp, err := httpClient.Do(req)
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
