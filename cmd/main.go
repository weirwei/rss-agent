package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/weirwei/rss-agent/internal/agent"
	"github.com/weirwei/rss-agent/internal/config"
	"github.com/weirwei/rss-agent/internal/constants"
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
	phFeishu := agent.NewPHFeishu(cfg.Feishu[constants.AgentTypePH])

	// 创建 ProductHunt 抓取器
	phFetcher := fetcher.NewPHFetcher()

	// 添加动态源
	if cfg.Fetcher.ProductHunt.Enabled {
		rssHelper.AddFeed(constants.AgentPH, phFetcher, config.FeedConfig{
			Dynamic:  true,
			Template: "https://decohack.com/producthunt-daily-{{date}}/",
			Format:   "2006-01-02",
		})
	}

	// 添加 RSS 源
	for _, rssCfg := range cfg.Fetcher.RSS {
		if rssCfg.Enabled {
			f := fetcher.NewRSSFetcher(nil)
			if rssCfg.Send {
				ag := agent.NewRSSFeishu(cfg.Feishu[constants.AgentTypeRSS])
				switch rssCfg.Name {
				case constants.AgentBestBlogs:
					ag.SetFormatter(agent.BestBlogsFormatter)
				}
				f = fetcher.NewRSSFetcher(ag)
			}

			rssHelper.AddFeed(rssCfg.Name, f, config.FeedConfig{
				URL: rssCfg.URL,
			})
		}
	}

	// 添加飞书代理
	agentHelper.AddAgent(constants.AgentTypePH, phFeishu, cfg.Feishu[constants.AgentTypePH].Cron)

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
