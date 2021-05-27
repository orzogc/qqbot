package google

import (
	"context"
	"time"

	"github.com/orzogc/qqbot/command/search/search_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
	googlesearch "github.com/rocketlaunchr/google-search"
)

const GoogleID = "google" // ID

// google搜索
type Google struct{}

// 搜索text，实现Search接口
func (g *Google) Search(text string) ([]search_utils.SearchResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(qqbot_utils.Timeout)*time.Second)
	defer cancel()

	result, err := googlesearch.Search(ctx, text)
	if err != nil {
		return nil, err
	}
	sr := make([]search_utils.SearchResult, 0, len(result))
	for _, r := range result {
		sr = append(sr, search_utils.SearchResult{
			URL:   r.URL,
			Title: r.Title,
		})
	}

	return sr, nil
}
