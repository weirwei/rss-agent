package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/weirwei/rss-agent/internal/log"
	"github.com/weirwei/rss-agent/internal/model"
)

// PHFetcher ProductHunt 页面获取器
type PHFetcher struct{}

// NewPHFetcher 创建ProductHunt获取器
func NewPHFetcher() *PHFetcher {
	return &PHFetcher{}
}

// Fetch 实现 FeedFetcher 接口 - ProductHunt方式
func (h *PHFetcher) Fetch(url string) (*model.FeedData, error) {
	log.Info("开始获取ProductHunt页面...")
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

	titles := titlePattern.FindAllStringSubmatch(content, -1)
	slogans := sloganPattern.FindAllStringSubmatch(content, -1)
	descs := descPattern.FindAllStringSubmatch(content, -1)
	links := linkPattern.FindAllStringSubmatch(content, -1)

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

	for i := 0; i < minLen; i++ {
		title := strings.TrimSpace(titles[i][1])
		slogan := strings.TrimSpace(slogans[i][1])
		desc := strings.TrimSpace(descs[i][1])
		link := strings.TrimSpace(links[i][1])

		title = regexp.MustCompile(`^\d+\.\s*`).ReplaceAllString(title, "")

		item := model.FeedItem{
			Title:       title,
			Link:        link,
			Published:   time.Now(),
			Summary:     slogan,
			Description: desc,
		}
		result.Items = append(result.Items, item)
	}
	log.Info("获取ProductHunt页面成功。标题：%s", result.Title)
	return result, nil
}

func (h *PHFetcher) Complete(data *model.FeedData) error {
	return nil
}
