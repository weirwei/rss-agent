package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/model"
)

// RSSHelper RSS助手服务
type RSSHelper struct {
	feeds     map[string]config.FeedConfig
	outputDir string
	stopChan  chan bool
}

// NewRSSHelper 创建新的RSS助手实例
func NewRSSHelper(outputDir string) *RSSHelper {
	if outputDir == "" {
		outputDir = "rss_output"
	}

	// 创建输出目录
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("创建输出目录失败: %v\n", err)
	}

	return &RSSHelper{
		feeds:     make(map[string]config.FeedConfig),
		outputDir: outputDir,
		stopChan:  make(chan bool),
	}
}

// AddFeed 添加源
func (r *RSSHelper) AddFeed(name string, config config.FeedConfig) {
	r.feeds[name] = config
}

// FetchAllFeeds 抓取所有源
func (r *RSSHelper) FetchAllFeeds() map[string]*model.FeedData {
	results := make(map[string]*model.FeedData)

	for name, config := range r.feeds {
		url := config.URL
		if config.Dynamic {
			currentDate := time.Now().Format(config.Format)
			url = strings.Replace(config.Template, "{{date}}", currentDate, -1)
		}

		if feed, err := config.Fetcher.Fetch(url); err == nil {
			results[name] = feed
			// 保存到JSON
			if data, err := json.MarshalIndent(feed, "", "  "); err == nil {
				outputFile := filepath.Join(r.outputDir, name+".json")
				os.WriteFile(outputFile, data, 0644)
			}
		} else {
			fmt.Printf("抓取源 %s 失败: %v\n", name, err)
		}
	}

	return results
}

// StartSchedule 启动定时任务
func (r *RSSHelper) StartSchedule(intervalMinutes int) {
	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	fmt.Printf("开始定时任务，间隔时间：%d分钟\n", intervalMinutes)

	for {
		select {
		case <-ticker.C:
			fmt.Println("执行定时抓取任务...")
			r.FetchAllFeeds()
		case <-r.stopChan:
			fmt.Println("停止定时任务")
			return
		}
	}
}

// Stop 停止定时任务
func (r *RSSHelper) Stop() {
	r.stopChan <- true
}
