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
	Title     string `json:"title"`
	Link      string `json:"link"`
	Published string `json:"published"`
	Summary   string `json:"summary"`
	Author    string `json:"author"`
}

// DynamicFeed 表示动态RSS源的结构
type DynamicFeed struct {
	URLTemplate string
	DateFormat  string
}

// RSSHelper RSS助手结构体
type RSSHelper struct {
	feeds        map[string]string
	dynamicFeeds map[string]DynamicFeed
	outputDir    string
	parser       *gofeed.Parser
	stopChan     chan bool
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
		feeds:        make(map[string]string),
		dynamicFeeds: make(map[string]DynamicFeed),
		outputDir:    outputDir,
		parser:       gofeed.NewParser(),
		stopChan:     make(chan bool),
	}
}

// AddFeed 添加RSS源
func (r *RSSHelper) AddFeed(name, url string) {
	r.feeds[name] = url
}

// AddDynamicFeed 添加动态RSS源
func (r *RSSHelper) AddDynamicFeed(name, urlTemplate, dateFormat string) {
	r.dynamicFeeds[name] = DynamicFeed{
		URLTemplate: urlTemplate,
		DateFormat:  dateFormat,
	}
}

// FetchFeed 抓取单个RSS源
func (r *RSSHelper) FetchFeed(name string) (*Feed, error) {
	url, exists := r.feeds[name]
	if !exists {
		return nil, fmt.Errorf("RSS源 %s 不存在", name)
	}

	feed, err := r.parser.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("解析RSS源失败: %v", err)
	}

	result := &Feed{
		Title:       feed.Title,
		Description: feed.Description,
		LastUpdated: time.Now(),
		Entries:     make([]Entry, 0),
	}

	for _, item := range feed.Items {
		entry := Entry{
			Title:     item.Title,
			Link:      item.Link,
			Published: item.Published,
			Summary:   item.Description,
		}
		if item.Author != nil {
			entry.Author = item.Author.Name
		}
		result.Entries = append(result.Entries, entry)
	}

	outputFile := filepath.Join(r.outputDir, name+".json")
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("JSON序列化失败: %v", err)
	}

	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return nil, fmt.Errorf("写入文件失败: %v", err)
	}

	return result, nil
}

// FetchDynamicPage 抓取和解析 HTML 页面
func (r *RSSHelper) FetchDynamicPage(url string) (*Feed, error) {
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

	// 创建基本的 Feed 结构
	result := &Feed{
		Title:       "ProductHunt Daily",
		Description: "Daily ProductHunt Updates",
		LastUpdated: time.Now(),
		Entries:     make([]Entry, 0),
	}

	// 调整正则表达式以匹配实际的HTML结构
	titlePattern := regexp.MustCompile(`<h2><a[^>]*>(\d+\.\s*[^<]+)</a></h2>`)
	descPattern := regexp.MustCompile(`<strong>介绍</strong>：([^<]+)<br`)
	linkPattern := regexp.MustCompile(`<h2><a href="([^"]+)"`)

	titles := titlePattern.FindAllStringSubmatch(content, -1)
	descs := descPattern.FindAllStringSubmatch(content, -1)
	links := linkPattern.FindAllStringSubmatch(content, -1)

	fmt.Printf("找到 %d 个标题, %d 个描述, %d 个链接\n", len(titles), len(descs), len(links))

	// 确保找到的标题和描述数量匹配
	minLen := len(titles)
	if len(descs) < minLen {
		minLen = len(descs)
	}
	if len(links) < minLen {
		minLen = len(links)
	}

	for i := 0; i < minLen; i++ {
		title := strings.TrimSpace(titles[i][1])
		desc := strings.TrimSpace(descs[i][1])
		link := strings.TrimSpace(links[i][1])

		// 清理标题中的序号
		title = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(title, "")

		entry := Entry{
			Title:   title,
			Summary: desc,
			Link:    link,
		}
		result.Entries = append(result.Entries, entry)
	}

	// 如果没有找到任何条目
	if len(result.Entries) == 0 {
		return nil, fmt.Errorf("页面中未找到任何产品信息")
	}

	return result, nil
}

// FetchAllFeeds 抓取所有RSS源
func (r *RSSHelper) FetchAllFeeds() map[string]*Feed {
	results := make(map[string]*Feed)

	// 处理静态源
	for name := range r.feeds {
		if feed, err := r.FetchFeed(name); err == nil {
			results[name] = feed
		} else {
			fmt.Printf("抓取静态源 %s 失败: %v\n", name, err)
		}
	}

	// 处理动态源
	for name, df := range r.dynamicFeeds {
		currentDate := time.Now().Format(df.DateFormat)
		url := strings.Replace(df.URLTemplate, "{{date}}", currentDate, -1)

		if feed, err := r.FetchDynamicPage(url); err == nil {
			results[name] = feed

			// 保存到JSON文件
			outputFile := filepath.Join(r.outputDir, name+".json")
			data, err := json.MarshalIndent(feed, "", "  ")
			if err == nil {
				os.WriteFile(outputFile, data, 0644)
			}
		} else {
			fmt.Printf("抓取动态源 %s 失败: %v\n", name, err)
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
	helper.AddDynamicFeed(
		"producthunt-daily",
		"https://decohack.com/producthunt-daily-{{date}}/",
		"2006-01-02",
	)

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
