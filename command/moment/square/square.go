package square

import (
	"fmt"
	"sync"

	"github.com/orzogc/qqbot/command/moment/moment_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/valyala/fastjson"
)

const (
	AcFunSquareID  = "acfunSquare"                                           // ID
	AcFunSquareURL = "https://api-new.app.acfun.cn/rest/app/feed/feedSquare" // AcFun动态广场API
	Num            = 10                                                      // 限制获取的动态数量
)

// AcFun动态广场
type AcFunSquare struct{}

// 获取动态，实现Moment接口
func (a *AcFunSquare) GetMoment() ([]moment_utils.Moment, error) {
	resp, err := qqbot_utils.Get(AcFunSquareURL, nil)
	if err != nil {
		return nil, err
	}

	var p fastjson.Parser
	v, err := p.ParseBytes(resp)
	if err != nil {
		return nil, err
	}
	if !v.Exists("result") || v.GetInt("result") != 0 {
		return nil, fmt.Errorf("获取AcFun动态广场失败")
	}

	list := v.GetArray("feedList")
	list = list[:Num]
	result := make([]moment_utils.Moment, len(list))
	var mu1 sync.Mutex
	var wg1 sync.WaitGroup
	for i, l := range list {
		wg1.Add(1)
		go func(i int, l *fastjson.Value) {
			defer wg1.Done()
			imgs := l.GetArray("moment", "imgs")
			images := make([][]byte, len(imgs))
			var mu2 sync.Mutex
			var wg2 sync.WaitGroup
			for j, img := range imgs {
				wg2.Add(1)
				go func(j int, img *fastjson.Value) {
					defer wg2.Done()
					resp, err := qqbot_utils.Get(string(img.GetStringBytes("originUrl")), nil)
					if err != nil {
						return
					}
					mu2.Lock()
					defer mu2.Unlock()
					images[j] = resp
				}(j, img)
			}
			wg2.Wait()
			mu1.Lock()
			defer mu1.Unlock()
			result[i] = moment_utils.Moment{
				URL:    string(l.GetStringBytes("shareUrl")),
				Author: string(l.GetStringBytes("user", "userName")),
				Text:   string(l.GetStringBytes("discoveryResourceFeedShowContent")),
				Images: images,
			}
		}(i, l)
	}
	wg1.Wait()

	return result, nil
}
