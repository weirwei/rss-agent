package service

import (
	"fmt"
	"os"

	"github.com/robfig/cron/v3"
	"github.com/weirwei/rss-agent/internal/agent"
)

// AgentHelper 消息发送助手服务
type AgentHelper struct {
	agents   map[string]AgentConfig
	inputDir string
	cron     *cron.Cron
}

// AgentConfig 代理配置
type AgentConfig struct {
	Agent agent.Agent
	Cron  string
}

// NewAgentHelper 创建新的发送助手实例
func NewAgentHelper(inputDir string) *AgentHelper {
	if inputDir == "" {
		inputDir = "rss_output"
	}

	return &AgentHelper{
		agents:   make(map[string]AgentConfig),
		inputDir: inputDir,
		cron:     cron.New(cron.WithSeconds()),
	}
}

// AddAgent 添加发送代理
func (a *AgentHelper) AddAgent(name string, agent agent.Agent, cronExpr string) {
	a.agents[name] = AgentConfig{
		Agent: agent,
		Cron:  cronExpr,
	}
}

// SendAll 发送所有消息
func (a *AgentHelper) SendAll() {
	for name, agentConfig := range a.agents {
		// 读取对应的数据文件
		file, err := os.ReadFile(fmt.Sprintf("%s/%s.json", a.inputDir, name))
		if err != nil {
			fmt.Printf("读取文件失败 %s: %v\n", name, err)
			continue
		}

		// 发送消息
		err = agentConfig.Agent.Send(fmt.Sprintf("%s Daily", name), file)
		if err != nil {
			fmt.Printf("发送消息失败 %s: %v\n", name, err)
		}
	}
}

// StartSchedule 启动定时任务
func (a *AgentHelper) StartSchedule() error {
	for name, agentConfig := range a.agents {
		agentName := name // 创建副本用于闭包
		agent := agentConfig.Agent

		_, err := a.cron.AddFunc(agentConfig.Cron, func() {
			fmt.Printf("执行定时发送任务: %s\n", agentName)
			file, err := os.ReadFile(fmt.Sprintf("%s/%s.json", a.inputDir, agentName))
			if err != nil {
				fmt.Printf("读取文件失败 %s: %v\n", agentName, err)
				return
			}

			err = agent.Send(fmt.Sprintf("%s Daily", agentName), file)
			if err != nil {
				fmt.Printf("发送消息失败 %s: %v\n", agentName, err)
			}
		})

		if err != nil {
			return fmt.Errorf("添加定时任务失败 %s: %v", name, err)
		}
	}

	a.cron.Start()
	return nil
}

// Stop 停止定时任务
func (a *AgentHelper) Stop() {
	if a.cron != nil {
		a.cron.Stop()
	}
}
