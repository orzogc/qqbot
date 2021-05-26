package acfun

import (
	"fmt"
	"net/url"

	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/search/search_utils"
	"github.com/valyala/fastjson"
)

const (
	AcFunID        = "acfun"                                                // ID
	AcFunSearchURL = "https://api-new.app.acfun.cn/rest/app/search/complex" // AcFun综合搜索API
	acfunUserPage  = "https://www.acfun.cn/u/%d"                            // AcFun个人页面
	acfunVedioPage = "https://www.acfun.cn/v/ac%d"                          // AcFun视频页面
	keyword        = "keyword"
)

// AcFun综合搜索
type AcFun struct{}

func (a *AcFun) Search(text string) ([]search_utils.SearchResult, error) {
	query := url.Values{}
	query.Add(keyword, text)
	resp, err := qqbot_utils.Get(AcFunSearchURL, query)
	if err != nil {
		return nil, err
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return nil, err
	}
	if !v.Exists("result") || v.GetInt("result") != 0 {
		return nil, fmt.Errorf("AcFun综合搜索失败：%s", text)
	}

	list := v.GetArray("itemList")
	result := make([]search_utils.SearchResult, 0, len(list))
	for _, l := range list {
		if l.Exists("contentId") {
			result = append(result, search_utils.SearchResult{
				URL:   fmt.Sprintf(acfunVedioPage, l.GetInt("contentId")),
				Title: string(l.GetStringBytes("userName")) + " " + string(l.GetStringBytes("title")),
			})
		} else {
			result = append(result, search_utils.SearchResult{
				URL:   fmt.Sprintf(acfunUserPage, l.GetInt("userId")),
				Title: string(l.GetStringBytes("userName")),
			})
		}
	}

	return result, nil
}
