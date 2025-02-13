package agent

import (
	"regexp"
	"time"

	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/model"
)

type rssFeishu struct {
	webhookURL string
	length     int
	formatter  DataFormatter
}

type DataFormatter func(*model.FeedData)

func NewRSSFeishu(config config.AgentConfig, dateFormatter ...DataFormatter) Agent {
	feishu := &rssFeishu{
		webhookURL: config.WebhookURL,
		length:     config.Length,
	}
	if len(dateFormatter) > 0 {
		feishu.formatter = dateFormatter[0]
	}

	return feishu
}

func (r *rssFeishu) Send(data model.FeedData) error {
	if r.formatter != nil {
		r.formatter(&data)
	}
	title, content, err := r.formatToMarkdown(data)
	if err != nil {
		return err
	}
	return SendToFeishu(r.webhookURL, title, content)
}

func (r *rssFeishu) SetFormatter(formatter DataFormatter) {
	r.formatter = formatter
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
			Text: item.Description + "\n\n",
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

var bestBlogsFormatterRe = regexp.MustCompile(`</h3>\s*<p[^>]*>([^<]*?)</p>`)

func BestBlogsFormatter(data *model.FeedData) {
	for i, v := range data.Items {
		matches := bestBlogsFormatterRe.FindAllStringSubmatch(v.Summary, -1)
		if len(matches) > 0 {
			data.Items[i].Summary = matches[0][1]
		}
		if len(matches) > 1 {
			data.Items[i].Description = matches[1][1]
		}
	}
}
