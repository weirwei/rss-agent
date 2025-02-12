package agent

const (
	AgentPHFeishu = "product_hunt"
)

type Agent interface {
	Send(data []byte) error
}
