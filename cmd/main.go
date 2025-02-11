package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/fetcher"
	"github.com/weirwei/rss-agent/internal/service"
)

func main() {
	helper := service.NewRSSHelper("")

	// 创建HTML抓取器
	htmlFetcher := fetcher.NewHTMLFetcher()

	// 添加动态源
	helper.AddFeed("producthunt-daily", config.FeedConfig{
		Fetcher:  htmlFetcher,
		Dynamic:  true,
		Template: "https://decohack.com/producthunt-daily-{{date}}/",
		Format:   "2006-01-02",
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
