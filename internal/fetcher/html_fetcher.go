package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/weirwei/rss-agent/internal/model"
)

// HTMLFetcher HTML页面获取器
type HTMLFetcher struct{}

// NewHTMLFetcher 创建HTML获取器
func NewHTMLFetcher() *HTMLFetcher {
	return &HTMLFetcher{}
}

// Fetch 实现 FeedFetcher 接口 - HTML方式
func (h *HTMLFetcher) Fetch(url string) (*model.FeedData, error) {
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
	result := &model.FeedData{
		Title:       "ProductHunt Daily",
		Description: "Daily ProductHunt Updates",
		LastUpdated: time.Now(),
		Items:       make([]model.FeedItem, 0),
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

		item := model.FeedItem{
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
