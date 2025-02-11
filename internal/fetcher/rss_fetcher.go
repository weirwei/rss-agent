package fetcher

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
	"github.com/weirwei/rss-agent/internal/model"
)

// RSSFetcher RSS源获取器
type RSSFetcher struct {
	parser *gofeed.Parser
}

// NewRSSFetcher 创建RSS获取器
func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{
		parser: gofeed.NewParser(),
	}
}

// Fetch 实现 FeedFetcher 接口 - RSS方式
func (r *RSSFetcher) Fetch(url string) (*model.FeedData, error) {
	feed, err := r.parser.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("解析RSS源失败: %v", err)
	}

	result := &model.FeedData{
		Title:       feed.Title,
		Description: feed.Description,
		LastUpdated: time.Now(),
		Items:       make([]model.FeedItem, 0),
	}

	for _, item := range feed.Items {
		pubTime := time.Now()
		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
		}

		feedItem := model.FeedItem{
			Title:       item.Title,
			Link:        item.Link,
			Published:   pubTime,
			Summary:     item.Description,
			Description: item.Content,
		}
		if item.Author != nil {
			feedItem.Author = item.Author.Name
		}
		result.Items = append(result.Items, feedItem)
	}

	return result, nil
}
