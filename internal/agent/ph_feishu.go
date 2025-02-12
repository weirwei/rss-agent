package agent

import (
	"encoding/json"
	"fmt"

	"github.com/weirwei/rss-agent/internal/config"
)

type phFeishu struct {
	webhookURL string
	length     int
}

func NewPHFeishu(config config.AgentConfig) *phFeishu {
	return &phFeishu{
		webhookURL: config.WebhookURL,
		length:     config.Length,
	}
}

func (p *phFeishu) Send(data []byte) error {
	title, content, err := p.formatToMarkdown(data)
	if err != nil {
		return err
	}
	return SendToFeishu(p.webhookURL, title, content)
}

func (p *phFeishu) formatToMarkdown(data []byte) (string, [][]interface{}, error) {
	var feed struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		LastUpdated string `json:"last_updated"`
		Items       []struct {
			Title       string `json:"title"`
			Link        string `json:"link"`
			Published   string `json:"published"`
			Summary     string `json:"summary"`
			Description string `json:"description"`
		} `json:"items"`
	}

	err := json.Unmarshal(data, &feed)
	if err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	var content [][]interface{}

	for i, item := range feed.Items {
		if i >= p.length {
			break
		}
		var row []interface{}

		row = append(row, TextElement{
			Tag:  "text",
			Text: "\n",
		})
		row = append(row, AElement{
			Tag:  "a",
			Text: item.Title + ": " + item.Summary,
			Href: item.Link,
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: "\n\n",
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: item.Description + "\n\n",
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: "--------------------------------\n",
		})

		content = append(content, row)
	}

	return feed.Title, content, nil
}
