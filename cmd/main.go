package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/weirwei/rss-agent/internal/agent"
	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/fetcher"
	"github.com/weirwei/rss-agent/internal/log"
	"github.com/weirwei/rss-agent/internal/service"
)

func main() {
	// 加载配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("加载配置失败: %v", err)
	}

	// 初始化 RSS 助手
	rssHelper := service.NewRSSHelper("")

	// 初始化 Agent 助手
	agentHelper := service.NewAgentHelper("")

	// 初始化飞书代理
	phFeishu := agent.NewPHFeishu(cfg.Feishu[agent.AgentPHFeishu])

	// 创建 ProductHunt 抓取器
	phFetcher := fetcher.NewPHFetcher()
	rssFetcher := fetcher.NewRSSFetcher()

	// 添加动态源
	if cfg.Fetcher.ProductHunt.Enabled {
		rssHelper.AddFeed("producthunt-daily", config.FeedConfig{
			Fetcher:  phFetcher,
			Dynamic:  true,
			Template: "https://decohack.com/producthunt-daily-{{date}}/",
			Format:   "2006-01-02",
		})
	}

	// 添加 RSS 源
	for _, rssCfg := range cfg.Fetcher.RSS {
		if rssCfg.Enabled {
			rssHelper.AddFeed(rssCfg.Name, config.FeedConfig{
				Fetcher: rssFetcher,
				URL:     rssCfg.URL,
			})
		}
	}

	// 添加飞书代理
	agentHelper.AddAgent("producthunt-daily", phFeishu, cfg.Feishu[agent.AgentPHFeishu].Cron)

	// 首次抓取和发送
	log.Info("开始首次抓取...")
	rssHelper.FetchAllFeeds()
	// agentHelper.SendAll()

	// 设置优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动抓取定时任务
	go rssHelper.StartSchedule(cfg.Fetcher.Interval)

	// 启动发送定时任务
	if err := agentHelper.StartSchedule(); err != nil {
		log.Fatal("启动发送定时任务失败: %v", err)
	}

	// 等待退出信号
	<-sigChan
	rssHelper.Stop()
	agentHelper.Stop()
	log.Info("程序已退出")
}
