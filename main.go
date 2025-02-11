package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"time"

	"github.com/mmcdole/gofeed"
)

// Feed 表示单个RSS源的结构
type Feed struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	LastUpdated time.Time `json:"last_updated"`
	Entries     []Entry   `json:"entries"`
}

// Entry 表示RSS条目的结构
type Entry struct {
	Title       string `json:"title"`
	Link        string `json:"link"`
	Published   string `json:"published"`
	Summary     string `json:"summary"`     // 用于存放标语
	Description string `json:"description"` // 用于存放介绍
	Author      string `json:"author"`
}

// DynamicFeed 表示动态RSS源的结构
type DynamicFeed struct {
	URLTemplate string
	DateFormat  string
}

// FeedFetcher 定义了获取数据的统一接口
type FeedFetcher interface {
	Fetch(url string) (*FeedData, error)
}

// FeedData 统一的数据结构
type FeedData struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	LastUpdated time.Time  `json:"last_updated"`
	Items       []FeedItem `json:"items"`
}

// FeedItem 统一的条目结构
type FeedItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Published   time.Time `json:"published"`
	Summary     string    `json:"summary"`     // 标语/简短描述
	Description string    `json:"description"` // 详细描述
	Author      string    `json:"author,omitempty"`
}

// RSSFetcher RSS源获取器
type RSSFetcher struct {
	parser *gofeed.Parser
}

// HTMLFetcher HTML页面获取器
type HTMLFetcher struct{}

// RSSHelper RSS助手结构体
type RSSHelper struct {
	feeds     map[string]FeedConfig
	outputDir string
	stopChan  chan bool
}

type FeedConfig struct {
	URL      string
	Fetcher  FeedFetcher
	Dynamic  bool
	Template string
	Format   string
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
		feeds:     make(map[string]FeedConfig),
		outputDir: outputDir,
		stopChan:  make(chan bool),
	}
}

// NewRSSFetcher 创建RSS获取器
func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{
		parser: gofeed.NewParser(),
	}
}

// NewHTMLFetcher 创建HTML获取器
func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{}
}

// Fetch 实现 FeedFetcher 接口 - RSS方式
func (r *RSSFetcher) Fetch(url string) (*FeedData, error) {
	feed, err := r.parser.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("解析RSS源失败: %v", err)
	}

	result := &FeedData{
		Title:       feed.Title,
		Description: feed.Description,
		LastUpdated: time.Now(),
		Items:       make([]FeedItem, 0),
	}

	for _, item := range feed.Items {
		pubTime := time.Now()
		if item.PublishedParsed != nil {
			pubTime = *item.PublishedParsed
		}

		feedItem := FeedItem{
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

// Fetch 实现 FeedFetcher 接口 - HTML方式
func (h *HTMLFetcher) Fetch(url string) (*FeedData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取页面失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP状态码错误: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取页面内容失败: %v", err)
	}

	content := string(body)
	result := &FeedData{
		Title:       "ProductHunt Daily",
		Description: "Daily ProductHunt Updates",
		LastUpdated: time.Now(),
		Items:       make([]FeedItem, 0),
	}

	// 解析HTML内容
	titlePattern := regexp.MustCompile(`<h2><a[^>]*>(\d+\.\s*[^<]+)</a></h2>`)
	sloganPattern := regexp.MustCompile(`<strong>标语</strong>：([^<]+)<br`)
	descPattern := regexp.MustCompile(`<strong>介绍</strong>：([^<]+)<br`)
	linkPattern := regexp.MustCompile(`<h2><a href="([^"]+)"`)
	websitePattern := regexp.MustCompile(`<strong>产品网站</strong>: <a href="([^"]+)"`)

	titles := titlePattern.FindAllStringSubmatch(content, -1)
	slogans := sloganPattern.FindAllStringSubmatch(content, -1)
	descs := descPattern.FindAllStringSubmatch(content, -1)
	links := linkPattern.FindAllStringSubmatch(content, -1)
	websites := websitePattern.FindAllStringSubmatch(content, -1)

	fmt.Printf("找到 %d 个标题, %d 个标语, %d 个描述, %d 个链接, %d 个网站\n",
		len(titles), len(slogans), len(descs), len(links), len(websites))

	// 确保找到的各项数量匹配
	minLen := len(titles)
	if len(slogans) < minLen {
		minLen = len(slogans)
	}
	if len(descs) < minLen {
		minLen = len(descs)
	}
	if len(links) < minLen {
		minLen = len(links)
	}
	if len(websites) < minLen {
		minLen = len(websites)
	}

	for i := 0; i < minLen; i++ {
		title := strings.TrimSpace(titles[i][1])
		slogan := strings.TrimSpace(slogans[i][1])
		desc := strings.TrimSpace(descs[i][1])
		link := strings.TrimSpace(links[i][1])
		website := strings.TrimSpace(websites[i][1])

		title = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(title, "")
		fullDesc := fmt.Sprintf("%s\n\n【产品网站】%s", desc, website)

		item := FeedItem{
			Title:       title,
			Link:        link,
			Published:   time.Now(),
			Summary:     slogan,
			Description: fullDesc,
		}
		result.Items = append(result.Items, item)
	}

	return result, nil
}

func (r *RSSHelper) AddFeed(name string, config FeedConfig) {
	r.feeds[name] = config
}

// FetchAllFeeds 抓取所有RSS源
func (r *RSSHelper) FetchAllFeeds() map[string]*FeedData {
	results := make(map[string]*FeedData)

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

func main() {
	// 创建RSS助手实例
	helper := NewRSSHelper("")

	// 添加RSS源
	// helper.AddFeed("zhihu", "https://www.zhihu.com/rss")
	// helper.AddFeed("decohack", "https://decohack.com/feed")

	// 添加动态RSS源
	helper.AddFeed("producthunt-daily", FeedConfig{
		URL:     "https://decohack.com/producthunt-daily-{{date}}/",
		Format:  "2006-01-02",
		Dynamic: true,
	})

	// 首次抓取
	fmt.Println("开始首次抓取...")
	results := helper.FetchAllFeeds()
	fmt.Printf("已完成首次抓取，获取到 %d 个源的数据\n", len(results))

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动定时任务（30分钟间隔）
	go helper.StartSchedule(30)

	// 等待退出信号
	<-sigChan
	helper.Stop()
	fmt.Println("程序已退出")
}
