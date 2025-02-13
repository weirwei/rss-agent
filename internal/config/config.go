package config

import (
	"github.com/spf13/viper"
	"github.com/weirwei/rss-agent/internal/fetcher"
)

// FeedConfig 表示一个源的配置
type FeedConfig struct {
	URL      string
	Fetcher  fetcher.FeedFetcher
	Dynamic  bool
	Template string
	Format   string
}

type Config struct {
	App     AppConfig              `mapstructure:"app"`
	Feishu  map[string]AgentConfig `mapstructure:"feishu"`
	Fetcher FetcherConfig          `mapstructure:"fetcher"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
}

type AgentConfig struct {
	WebhookURL string `mapstructure:"webhook_url"`
	Cron       string `mapstructure:"cron"`
	Length     int    `mapstructure:"length"`
}

type FetcherConfig struct {
	Interval    int               `mapstructure:"interval"`
	ProductHunt ProductHuntConfig `mapstructure:"product_hunt"`
	RSS         []RSSConfig       `mapstructure:"rss"`
}

type ProductHuntConfig struct {
	Enabled bool `mapstructure:"enabled"`
	Length  int  `mapstructure:"length"`
}

type RSSConfig struct {
	Name    string `mapstructure:"name"`
	URL     string `mapstructure:"url"`
	Enabled bool   `mapstructure:"enabled"`
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
