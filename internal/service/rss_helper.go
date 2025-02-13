package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/constants"
	"github.com/weirwei/rss-agent/internal/fetcher"
	"github.com/weirwei/rss-agent/internal/log"
	"github.com/weirwei/rss-agent/internal/model"
)

// RSSHelper RSS助手服务
type RSSHelper struct {
	feeds     map[constants.AgentName]config.FeedConfig
	fetchers  map[constants.AgentName]fetcher.FeedFetcher
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
		log.Error("创建输出目录失败: %v", err)
	}

	return &RSSHelper{
		feeds:     make(map[constants.AgentName]config.FeedConfig),
		fetchers:  make(map[constants.AgentName]fetcher.FeedFetcher),
		outputDir: outputDir,
		stopChan:  make(chan bool),
	}
}

// AddFeed 添加源
func (r *RSSHelper) AddFeed(name constants.AgentName, fetcher fetcher.FeedFetcher, config config.FeedConfig) {
	r.feeds[name] = config
	r.fetchers[name] = fetcher
}

// FetchAllFeeds 抓取所有源
func (r *RSSHelper) FetchAllFeeds() {
	for name, config := range r.feeds {
		url := config.URL
		if config.Dynamic {
			currentDate := time.Now().Format(config.Format)
			url = strings.Replace(config.Template, "{{date}}", currentDate, -1)
		}
		f, ok := r.fetchers[name]
		if !ok {
			log.Error("未找到 fetcher %s", name)
			continue
		}
		if feed, err := f.Fetch(url); err == nil {
			var oldFeed model.FeedData
			file, err := os.ReadFile(fmt.Sprintf("%s/%s.json", r.outputDir, name))
			if err != nil && !os.IsNotExist(err) {
				log.Error("读取文件失败 %s: %v", name, err)
				continue
			}
			if len(file) > 0 {
				err = jsoniter.Unmarshal(file, &oldFeed)
				if err != nil {
					log.Error("解析旧数据失败 %s: %v", name, err)
				}
			}
			// 最后更新时间相同，不更新
			if feed.LastUpdated.Equal(oldFeed.LastUpdated) {
				continue
			}
			// 保存到JSON
			if data, err := json.MarshalIndent(feed, "", "  "); err == nil {
				outputFile := filepath.Join(r.outputDir, string(name)+".json")
				os.WriteFile(outputFile, data, 0644)
			}
			// 用增量数据执行后处理
			latestFeed := fetcher.LatestFeed(oldFeed, *feed)
			err = f.Complete(&latestFeed)
			if err != nil {
				log.Error("完成抓取失败 %s: %v", name, err)
				continue
			}
		} else {
			log.Error("抓取源 %s 失败: %v", name, err)
		}
	}

}

// StartSchedule 启动定时任务
func (r *RSSHelper) StartSchedule(intervalMinutes int) {
	if intervalMinutes <= 0 {
		log.Error("定时任务间隔必须大于0分钟")
		return
	}

	ticker := time.NewTicker(time.Duration(intervalMinutes) * time.Minute)
	defer ticker.Stop()

	log.Info("开始定时任务，间隔时间：%d分钟", intervalMinutes)

	for {
		select {
		case <-ticker.C:
			log.Info("执行定时抓取任务...")
			r.FetchAllFeeds()
		case <-r.stopChan:
			log.Info("停止定时任务")
			return
		}
	}
}

// Stop 停止定时任务
func (r *RSSHelper) Stop() {
	r.stopChan <- true
}
