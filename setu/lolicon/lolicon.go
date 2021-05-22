package lolicon

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
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

type Response struct {
	Code        Code     `json:"code"`
	Msg         string   `json:"msg"`
	Quota       int      `json:"quota"`
	QuotaMinTTL int      `json:"quota_min_ttl"`
	Count       int      `json:"count"`
	Data        []*Image `json:"data"`
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

type Query struct {
	APIKey   string `json:"apikey"`
	R18      uint   `json:"r18"`
	Keyword  string `json:"keyword"`
	Num      uint   `json:"num"`
	Proxy    string `json:"proxy"`
	Size1200 bool   `json:"size1200"`
}

var client = &http.Client{
	Timeout: 20 * time.Second,
}

func SetTimeout(second uint) {
	client = &http.Client{
		Timeout: time.Duration(second) * time.Second,
	}
}

func (r *Response) IsSuccess() bool {
	return r.Code == Success
}

func (r *Response) GetImage() ([][]byte, error) {
	if !r.IsSuccess() {
		return nil, fmt.Errorf("请求lolicon失败，返回码code为%d，错误信息msg为%s", r.Code, r.Msg)
	}
	if len(r.Data) == 0 {
		return nil, fmt.Errorf("请求lolicon失败，data长度为0")
	}

	images := make([][]byte, 0, len(r.Data))
	var mu sync.Mutex
	var wg sync.WaitGroup
	for _, img := range r.Data {
		wg.Add(1)
		go func(img *Image) {
			defer wg.Done()
			req, err := http.NewRequest(http.MethodGet, img.URL, nil)
			if err != nil {
				return
			}
			if req.URL.Host == "i.pximg.net" {
				req.Header.Set("Referer", "https://www.pixiv.net/")
			}
			resp, err := client.Do(req)
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
		return nil, fmt.Errorf("获取图片失败")
	}

	return images, nil
}

func (q *Query) Lolicon() (*Response, error) {
	req, err := http.NewRequest(http.MethodGet, LoliconURL, nil)
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	if q.APIKey != "" {
		query.Add(APIKey, q.APIKey)
	}
	if q.R18 > 2 {
		return nil, fmt.Errorf("R18必须为0、1或2，现为%d", q.R18)
	}
	if q.R18 != 0 {
		query.Add(R18, strconv.FormatUint(uint64(q.R18), 10))
	}
	if q.Keyword != "" {
		query.Add(Keyword, q.Keyword)
	}
	if q.Num > 10 {
		return nil, fmt.Errorf("Num必须为0到10，现为%d", q.Num)
	}
	if q.Num != 0 {
		query.Add(Num, strconv.FormatUint(uint64(q.Num), 10))
	}
	if q.Proxy != "" {
		query.Add(Proxy, q.Proxy)
	}
	if q.Size1200 {
		query.Add(Size1200, "1")
	}
	req.URL.RawQuery = query.Encode()

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	r := new(Response)
	if err = json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}

func (q *Query) GetImage() ([][]byte, error) {
	r, err := q.Lolicon()
	if err != nil {
		return nil, err
	}

	return r.GetImage()
}
