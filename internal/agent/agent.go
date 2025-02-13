package agent

import "github.com/weirwei/rss-agent/internal/model"

type Agent interface {
	Send(data model.FeedData) error
	SetFormatter(formatter DataFormatter)
}
