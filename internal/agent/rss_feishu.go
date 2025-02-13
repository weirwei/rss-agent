package agent

import (
	"time"

	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/model"
)

type rssFeishu struct {
	webhookURL string
	length     int
}

func NewRSSFeishu(config config.AgentConfig) *rssFeishu {
	return &rssFeishu{
		webhookURL: config.WebhookURL,
		length:     config.Length,
	}
}

func (r *rssFeishu) Send(data model.FeedData) error {
	title, content, err := r.formatToMarkdown(data)
	if err != nil {
		return err
	}
	return SendToFeishu(r.webhookURL, title, content)
}

func (r *rssFeishu) formatToMarkdown(data model.FeedData) (string, [][]interface{}, error) {
	var content [][]interface{}
	for i, item := range data.Items {
		if i >= r.length {
			break
		}
		var row []interface{}

		row = append(row, TextElement{
			Tag:  "text",
			Text: "\n",
		})
		row = append(row, AElement{
			Tag:  "a",
			Text: item.Title,
			Href: item.Link,
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: "\n\n",
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: item.Summary + "\n\n",
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: "发布时间：" + item.Published.Format(time.DateTime) + "\n\n",
		})
		row = append(row, TextElement{
			Tag:  "text",
			Text: "--------------------------------\n",
		})

		content = append(content, row)
	}

	return data.Title, content, nil
}
