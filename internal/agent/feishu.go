package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/weirwei/rss-agent/internal/log"
)

type TextElement struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

type AElement struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
	Href string `json:"href"`
}

// FeishuMessage represents the message structure for Feishu
type FeishuMessage struct {
	MsgType string `json:"msg_type"`
	Content struct {
		Post struct {
			ZhCn struct {
				Title   string          `json:"title"`
				Content [][]interface{} `json:"content"`
			} `json:"zh_cn"`
		} `json:"post"`
	} `json:"content"`
}

// SendToFeishu sends a message to the Feishu robot
func SendToFeishu(feishuWebhookURL string, title string, content [][]interface{}) error {
	msg := FeishuMessage{
		MsgType: "post",
		Content: struct {
			Post struct {
				ZhCn struct {
					Title   string          `json:"title"`
					Content [][]interface{} `json:"content"`
				} `json:"zh_cn"`
			} `json:"post"`
		}{
			Post: struct {
				ZhCn struct {
					Title   string          `json:"title"`
					Content [][]interface{} `json:"content"`
				} `json:"zh_cn"`
			}{
				ZhCn: struct {
					Title   string          `json:"title"`
					Content [][]interface{} `json:"content"`
				}{
					Title:   title,
					Content: content,
				},
			},
		},
	}
	jsonValue, _ := json.Marshal(msg)
	resp, err := http.Post(feishuWebhookURL, "application/json", bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("failed to send message to Feishu: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("feishu API error: %s", string(body))
	}
	log.Info("Message sent to Feishu successfully. Title:%s", title)
	return nil
}
