package lolicon

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"

	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/setu/setu_utils"
)

const (
	ID         = "lolicon"
	LoliconURL = "https://api.lolicon.app/setu"
	APIKey     = "apikey"
	R18        = "r18"
	Keyword    = "keyword"
	Num        = "num"
	Proxy      = "proxy"
	Size1200   = "size1200"
)

type Lolicon struct {
	APIKey   string `json:"apikey"`
	R18      uint   `json:"r18"`
	Keyword  string `json:"keyword"`
	Num      uint   `json:"num"`
	Proxy    string `json:"proxy"`
	Size1200 bool   `json:"size1200"`
}

type Response struct {
	Code        Code    `json:"code"`
	Msg         string  `json:"msg"`
	Quota       int     `json:"quota"`
	QuotaMinTTL int     `json:"quota_min_ttl"`
	Count       int     `json:"count"`
	Data        []Image `json:"data"`
}

type Image struct {
	PID    int      `json:"pid"`
	P      int      `json:"p"`
	UID    int      `json:"uid"`
	Title  string   `json:"title"`
	Author string   `json:"author"`
	URL    string   `json:"url"`
	R18    bool     `json:"r18"`
	Width  int      `json:"width"`
	Height int      `json:"height"`
	Tags   []string `json:"tags"`
}

type Code int

const (
	InternalError   Code = -1
	Success         Code = 0
	APIKeyError     Code = 401
	Refuse          Code = 403
	KeywordNotFound Code = 404
	QuotaLimit      Code = 429
)

var (
	ErrorInternal        = errors.New("lolicon内部错误")
	ErrorAPIKey          = errors.New("apikey不存在或被封禁")
	ErrorRefuse          = errors.New("由于不规范的操作而被拒绝调用")
	ErrorKeywordNotFound = errors.New("找不到符合关键字的图片")
	ErrorQuotaLimit      = errors.New("达到调用额度限制")
)

func (l *Lolicon) Lolicon() (*Response, error) {
	query := url.Values{}
	if l.APIKey != "" {
		query.Add(APIKey, l.APIKey)
	}
	if l.R18 != 0 {
		if l.R18 > 2 {
			return nil, fmt.Errorf("R18必须为0、1或2，现为%d", l.R18)
		}
		query.Add(R18, strconv.FormatUint(uint64(l.R18), 10))
	}
	if l.Keyword != "" {
		query.Add(Keyword, l.Keyword)
	}
	if l.Num != 0 {
		if l.Num > 10 {
			return nil, fmt.Errorf("Num必须为0到10，现为%d", l.Num)
		}
		query.Add(Num, strconv.FormatUint(uint64(l.Num), 10))
	}
	if l.Proxy != "" {
		query.Add(Proxy, l.Proxy)
	}
	if l.Size1200 {
		query.Add(Size1200, "1")
	}

	body, err := qqbot_utils.Get(LoliconURL, query)
	if err != nil {
		return nil, err
	}

	r := new(Response)
	if err = json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (l *Lolicon) GetImage(keyword string) ([][]byte, error) {
	lolicon := *l
	lolicon.Keyword = keyword
	r, err := lolicon.Lolicon()
	if err != nil {
		return nil, err
	}

	return r.GetImage()
}

func (r *Response) IsSuccess() bool {
	return r.Code == Success
}

func (r *Response) Error() error {
	switch r.Code {
	case InternalError:
		return ErrorInternal
	case Success:
		return nil
	case APIKeyError:
		return ErrorAPIKey
	case Refuse:
		return ErrorRefuse
	case KeywordNotFound:
		return ErrorKeywordNotFound
	case QuotaLimit:
		return ErrorQuotaLimit
	default:
		return fmt.Errorf("未知的错误码：%d", r.Code)
	}
}

func (r *Response) GetImage() ([][]byte, error) {
	if !r.IsSuccess() {
		return nil, fmt.Errorf("请求lolicon出现错误：%w，错误信息msg：%s", r.Error(), r.Msg)
	}
	if len(r.Data) == 0 {
		return nil, fmt.Errorf("请求lolicon失败，data长度为0")
	}

	images := make([][]byte, 0, len(r.Data))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, img := range r.Data {
		wg.Add(1)
		go func(img Image) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodGet, img.URL, nil)
			if err != nil {
				return
			}
			if req.URL.Host == "i.pximg.net" {
				req.Header.Set("Referer", setu_utils.PixivURL)
			}
			resp, err := qqbot_utils.Client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return
			}
			mu.Lock()
			defer mu.Unlock()
			images = append(images, body)
		}(img)
	}
	wg.Wait()

	if len(images) == 0 {
		return nil, fmt.Errorf("获取lolicon图片失败")
	}

	return images, nil
}
