package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Notifier struct {
	token string
}

func NewNotifier(token string) *Notifier {
	return &Notifier{token: token}
}

func (n *Notifier) SendText(chatID int64, text string, webAppURL string) error {
	if n.token == "" {
		return nil
	}
	payload := map[string]any{
		"chat_id": chatID,
		"text":    text,
		"reply_markup": map[string]any{
			"inline_keyboard": [][]map[string]any{
				{{
					"text":    "Открыть CLUTCH",
					"web_app": map[string]string{"url": webAppURL},
				}},
			},
		},
	}
	body, _ := json.Marshal(payload)
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", n.token)
	resp, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("telegram sendMessage status %d", resp.StatusCode)
	}
	return nil
}
