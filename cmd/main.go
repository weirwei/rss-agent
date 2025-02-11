package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	internalAgent "github.com/weirwei/rss-agent/internal/agent"
	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/fetcher"
	"github.com/weirwei/rss-agent/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化飞书代理
	feishuAgent := internalAgent.NewPHFeishu(cfg.Feishu[internalAgent.AgentPHFeishu])

	helper := service.NewRSSHelper("")

	// 创建HTML抓取器
	phFetcher := fetcher.NewPHFetcher()

	// 添加动态源
	helper.AddFeed("producthunt-daily", config.FeedConfig{
		Fetcher:  phFetcher,
		Dynamic:  true,
		Template: "https://decohack.com/producthunt-daily-{{date}}/",
		Format:   "2006-01-02",
	})

	// 首次抓取
	fmt.Println("开始首次抓取...")
	helper.FetchAllFeeds()

	// 读取 rss_output/producthunt-daily.json
	file, err := os.ReadFile("rss_output/producthunt-daily.json")
	if err != nil {
		fmt.Printf("Failed to read rss_output/producthunt-daily.json: %v\n", err)
		return
	}

	// 发送到 Feishu
	err = feishuAgent.Send("ProductHunt Daily", file)
	if err != nil {
		fmt.Printf("Failed to send to Feishu: %v\n", err)
		return
	}

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
