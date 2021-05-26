package duckduckgo

import (
	"bytes"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/search/search_utils"
)

const (
	DuckDuckGoID  = "duckduckgo"                       // ID
	DuckDuckGoURL = "https://html.duckduckgo.com/html" // DuckDuckGo网页
	q             = "q"
)

// DuckDuckGo搜索
type DuckDuckGo struct{}

// 搜索text，实现Search接口
func (d *DuckDuckGo) Search(text string) (result []search_utils.SearchResult, err error) {
	query := url.Values{}
	query.Add(q, text)
	resp, err := qqbot_utils.Get(DuckDuckGoURL, query)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return nil, err
	}
	doc.Find(".results .web-result").Each(func(i int, s *goquery.Selection) {
		link := s.Find(".result__a")
		if l, ok := link.Attr("href"); ok {
			result = append(result, search_utils.SearchResult{
				URL:   l,
				Title: link.Text(),
			})
		}
	})

	return result, nil
}
