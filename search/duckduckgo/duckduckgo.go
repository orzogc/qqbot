package duckduckgo

import (
	"github.com/orzogc/qqbot/search/search_utils"
	"github.com/sap-nocops/duckduckgogo/client"
)

const DuckDuckGoID = "duckduckgo" // ID

// DuckDuckGo搜索
type DuckDuckGo struct {
	client *client.DuckDuckGoSearchClient
}

// 新建Duckduckgo
func New() *DuckDuckGo {
	return &DuckDuckGo{
		client: client.NewDuckDuckGoSearchClient(),
	}
}

func (d *DuckDuckGo) Search(text string) ([]search_utils.SearchResult, error) {
	result, err := d.client.Search(text)
	if err != nil {
		return nil, err
	}
	sr := make([]search_utils.SearchResult, 0, len(result))
	for _, r := range result {
		sr = append(sr, search_utils.SearchResult{
			URL:   r.FormattedUrl,
			Title: r.Title,
		})
	}

	return sr, nil
}
