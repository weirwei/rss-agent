package model

import "time"

// FeedData 统一的数据结构
type FeedData struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	LastUpdated time.Time  `json:"last_updated"`
	Items       []FeedItem `json:"items"`
}

// FeedItem 统一的条目结构
type FeedItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Published   time.Time `json:"published"`
	Summary     string    `json:"summary"`     // 标语/简短描述
	Description string    `json:"description"` // 详细描述
	Author      string    `json:"author,omitempty"`
}
