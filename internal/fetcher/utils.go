package fetcher

import "github.com/weirwei/rss-agent/internal/model"

// LatestFeed 获取最新的文章
func LatestFeed(oldFeed, newFeed model.FeedData) model.FeedData {
	if len(oldFeed.Items) == 0 {
		return newFeed
	}
	repeatM := make(map[string]bool)
	for _, v := range oldFeed.Items {
		repeatM[v.Title] = true
	}
	latestFeedData := model.FeedData{
		Title:       newFeed.Title,
		Description: newFeed.Description,
		LastUpdated: newFeed.LastUpdated,
	}
	for _, v := range newFeed.Items {
		if repeatM[v.Title] {
			break
		}
		latestFeedData.Items = append(latestFeedData.Items, v)
	}
	return latestFeedData
}

// itemsEqual 比较两个 FeedItem 是否相等
func itemsEqual(a, b model.FeedItem) bool {
	return a.Title == b.Title &&
		a.Link == b.Link &&
		a.Published.Equal(b.Published) &&
		a.Summary == b.Summary &&
		a.Description == b.Description &&
		a.Author == b.Author
}
