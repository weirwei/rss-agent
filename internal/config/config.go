package config

import "github.com/weirwei/rss-agent/internal/fetcher"

// FeedConfig 表示一个源的配置
type FeedConfig struct {
	URL      string
	Fetcher  fetcher.FeedFetcher
	Dynamic  bool
	Template string
	Format   string
}
