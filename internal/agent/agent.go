package agent

import "github.com/weirwei/rss-agent/internal/model"

const (
	AgentPHFeishu = "producthunt-daily"
	AgentRSS      = "RSS"
)

type Agent interface {
	Send(data model.FeedData) error
	SetFormatter(formatter DataFormatter)
}
