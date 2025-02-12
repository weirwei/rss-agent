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

	// 初始化 RSS 助手
	rssHelper := service.NewRSSHelper("")

	// 初始化 Agent 助手
	agentHelper := service.NewAgentHelper("")

	// 初始化飞书代理
	phFeishu := internalAgent.NewPHFeishu(cfg.Feishu[internalAgent.AgentPHFeishu])

	// 创建 ProductHunt 抓取器
	phFetcher := fetcher.NewPHFetcher()

	// 添加动态源
	rssHelper.AddFeed("producthunt-daily", config.FeedConfig{
		Fetcher:  phFetcher,
		Dynamic:  true,
		Template: "https://decohack.com/producthunt-daily-{{date}}/",
		Format:   "2006-01-02",
	})

	// 添加飞书代理
	agentHelper.AddAgent("producthunt-daily", phFeishu, cfg.Feishu[internalAgent.AgentPHFeishu].Cron)

	// 首次抓取和发送
	fmt.Println("开始首次抓取...")
	rssHelper.FetchAllFeeds()
	agentHelper.SendAll()

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动抓取定时任务
	go rssHelper.StartSchedule(cfg.Fetcher.Schedule)

	// 启动发送定时任务
	if err := agentHelper.StartSchedule(); err != nil {
		log.Fatalf("启动发送定时任务失败: %v", err)
	}

	// 等待退出信号
	<-sigChan
	rssHelper.Stop()
	agentHelper.Stop()
	fmt.Println("程序已退出")
}
