package agent

const (
	AgentPHFeishu = "feishu_ph"
)

type Agent interface {
	Send(title string, data []byte) error
}
