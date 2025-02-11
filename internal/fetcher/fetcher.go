package fetcher

import "github.com/weirwei/rss-agent/internal/model"

// FeedFetcher 定义了获取数据的统一接口
type FeedFetcher interface {
	Fetch(url string) (*model.FeedData, error)
}
