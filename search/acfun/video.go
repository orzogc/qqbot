package acfun

import (
	"fmt"
	"net/url"

	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/search/search_utils"
	"github.com/valyala/fastjson"
)

const (
	AcFunVideoID        = "acfun"                                              // ID
	AcFunVideoSearchURL = "https://api-new.app.acfun.cn/rest/app/search/video" // AcFun视频搜索API
	acfunVedioPage      = "https://www.acfun.cn/v/ac%d"                        // AcFun视频页面
	keyword             = "keyword"
	sortType            = "sortType"
	uploadTime          = "5"
)

// AcFun视频搜索
type AcFunVideo struct{}

func (a *AcFunVideo) Search(text string) ([]search_utils.SearchResult, error) {
	query := url.Values{}
	query.Add(keyword, text)
	query.Add(sortType, uploadTime)
	resp, err := qqbot_utils.Get(AcFunVideoSearchURL, query)
	if err != nil {
		return nil, err
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return nil, err
	}
	if !v.Exists("result") || v.GetInt("result") != 0 {
		return nil, fmt.Errorf("AcFun视频搜索失败：%s", text)
	}

	list := v.GetArray("videoList")
	result := make([]search_utils.SearchResult, 0, len(list))
	for _, l := range list {
		result = append(result, search_utils.SearchResult{
			URL:   fmt.Sprintf(acfunVedioPage, l.GetInt("contentId")),
			Title: string(l.GetStringBytes("userName")) + " " + string(l.GetStringBytes("title")),
		})
	}

	return result, nil
}
