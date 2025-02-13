package agent

import (
	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/model"
)

type phFeishu struct {
	webhookURL string
	length     int
}

func NewPHFeishu(config config.AgentConfig) Agent {
	return &phFeishu{
		webhookURL: config.WebhookURL,
		length:     config.Length,
	}
}

func (p *phFeishu) Send(data model.FeedData) error {
	content, err := p.formatToMarkdown(data)
	if err != nil {
		return err
	}
	return SendToFeishu(p.webhookURL, data.Title, content)
}

func (p *phFeishu) SetFormatter(formatter DataFormatter) {
}

func (p *phFeishu) formatToMarkdown(data model.FeedData) ([][]interface{}, error) {
	var content [][]interface{}
	for i, item := range data.Items {
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

	return content, nil
}
