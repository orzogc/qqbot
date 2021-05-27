package ehentai

import (
	"bytes"
	"net/url"

	"github.com/PuerkitoBio/goquery"
	"github.com/orzogc/qqbot/command/search/search_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	EHentaiID  = "ehentai"               // ID
	EHentaiURL = "https://e-hentai.org/" // E-Hentai网页
	f_search   = "f_search"
)

// E-Hentai
type EHentai struct{}

// 搜索text，实现Search接口
func (e *EHentai) Search(text string) (result []search_utils.SearchResult, err error) {
	query := url.Values{}
	query.Add(f_search, text)
	resp, err := qqbot_utils.Get(EHentaiURL, query)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(resp))
	if err != nil {
		return nil, err
	}
	doc.Find(".gl3c a").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".glink")
		if l, ok := s.Attr("href"); ok {
			result = append(result, search_utils.SearchResult{
				URL:   l,
				Title: title.Text(),
			})
		}
	})

	return result, nil
}
