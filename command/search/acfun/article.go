package acfun

import (
	"fmt"
	"net/url"

	"github.com/orzogc/qqbot/command/search/search_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/valyala/fastjson"
)

const (
	AcFunArticleID        = "acfunArticle"                                         // ID
	AcFunArticleSearchURL = "https://api-new.app.acfun.cn/rest/app/search/article" // AcFun文章搜索API
	acfunArticlePage      = "https://www.acfun.cn/a/ac%d"                          // AcFun文章页面
)

// AcFun文章搜索
type AcFunArticle struct{}

// 搜索text，实现Search接口
func (a *AcFunArticle) Search(text string) ([]search_utils.SearchResult, error) {
	query := url.Values{}
	query.Add(keyword, text)
	query.Add(sortType, uploadTime)
	resp, err := qqbot_utils.Get(AcFunArticleSearchURL, query)
	if err != nil {
		return nil, err
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return nil, err
	}
	if !v.Exists("result") || v.GetInt("result") != 0 {
		return nil, fmt.Errorf("AcFun文章搜索失败：%s", text)
	}

	list := v.GetArray("articleList")
	result := make([]search_utils.SearchResult, 0, len(list))
	for _, l := range list {
		result = append(result, search_utils.SearchResult{
			URL:   fmt.Sprintf(acfunArticlePage, l.GetInt("contentId")),
			Title: string(l.GetStringBytes("userName")) + " " + string(l.GetStringBytes("title")),
		})
	}

	return result, nil
}
