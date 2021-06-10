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

	"github.com/orzogc/qqbot/command/setu/setu_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	LoliconID  = "lolicon"                      // ID
	LoliconURL = "https://api.lolicon.app/setu" // lolicon接口
	apikey     = "apikey"
	r18        = "r18"
	keyword    = "keyword"
	num        = "num"
	proxy      = "proxy"
	size1200   = "size1200"
)

// lolicon的请求query
type Lolicon struct {
	APIKey   string `json:"apikey"`   // lolicon接口的apikey，需自行申请
	R18      uint   `json:"r18"`      // 0为非 R18，1为 R18，2为混合
	Keyword  string `json:"keyword"`  // 若指定关键字，将会返回从插画标题、作者、标签中模糊搜索的结果
	Num      uint   `json:"num"`      // 一次返回的结果数量，范围为1到10，不提供 APIKEY 时固定为1；在指定关键字的情况下，结果数量可能会不足指定的数量
	Proxy    string `json:"proxy"`    // 设置返回的原图链接的域名，设置为disable返回真正的原图链接
	Size1200 bool   `json:"size1200"` // 是否使用 master_1200 缩略图，即长或宽最大为 1200px 的缩略图，以节省流量或提升加载速度（某些原图的大小可以达到十几MB）
}

// lolicon请求的响应
type Response struct {
	Code        Code    `json:"code"`          // 返回码
	Msg         string  `json:"msg"`           // 错误信息
	Quota       int     `json:"quota"`         // 剩余调用额度
	QuotaMinTTL int     `json:"quota_min_ttl"` // 距离下一次调用额度恢复(+1)的秒数
	Count       int     `json:"count"`         // 结果数
	Data        []Image `json:"data"`          // 图片数组
}

// 图片
type Image struct {
	PID    int      `json:"pid"`    // 作品 PID
	P      int      `json:"p"`      // 作品所在 P
	UID    int      `json:"uid"`    // 作者 UID
	Title  string   `json:"title"`  // 作品标题
	Author string   `json:"author"` // 作者名（入库时，并过滤掉 @ 及其后内容）
	URL    string   `json:"url"`    // 图片链接（可能存在有些作品因修改或删除而导致 404 的情况）
	R18    bool     `json:"r18"`    // 是否 R18（在图库中的分类，并非作者标识的 R18）
	Width  int      `json:"width"`  // 原图宽度 px
	Height int      `json:"height"` // 原图高度 px
	Tags   []string `json:"tags"`   // 作品标签，包含标签的中文翻译（有的话）
}

// 返回码
type Code int

const (
	InternalError   Code = -1  // 内部错误
	Success         Code = 0   // 成功
	APIKeyError     Code = 401 // APIKEY 不存在或被封禁
	Refuse          Code = 403 // 由于不规范的操作而被拒绝调用
	KeywordNotFound Code = 404 // 找不到符合关键字的图片
	QuotaLimit      Code = 429 // 达到调用额度限制
)

var (
	ErrorInternal        = errors.New("lolicon内部错误")
	ErrorAPIKey          = errors.New("apikey不存在或被封禁")
	ErrorRefuse          = errors.New("由于不规范的操作而被拒绝调用")
	ErrorKeywordNotFound = errors.New("找不到符合关键字的图片")
	ErrorQuotaLimit      = errors.New("达到调用额度限制")
)

// 获取lolicon图片
func (l *Lolicon) Lolicon() (*Response, error) {
	query := url.Values{}
	if l.APIKey != "" {
		query.Add(apikey, l.APIKey)
	}
	if l.R18 != 0 {
		if l.R18 > 2 {
			return nil, fmt.Errorf("R18必须为0、1或2，现为%d", l.R18)
		}
		query.Add(r18, strconv.FormatUint(uint64(l.R18), 10))
	}
	if l.Keyword != "" {
		query.Add(keyword, l.Keyword)
	}
	if l.Num != 0 {
		if l.Num > 10 {
			return nil, fmt.Errorf("Num必须为0到10，现为%d", l.Num)
		}
		query.Add(num, strconv.FormatUint(uint64(l.Num), 10))
	}
	if l.Proxy != "" {
		query.Add(proxy, l.Proxy)
	}
	if l.Size1200 {
		query.Add(size1200, "1")
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

// 获取图片，实现Setu接口
func (l *Lolicon) GetImage(keyword string) (*setu_utils.Image, error) {
	lolicon := *l
	lolicon.Keyword = keyword
	r, err := lolicon.Lolicon()
	if err != nil {
		return nil, err
	}
	img, err := r.GetImage()
	if err != nil {
		return nil, err
	}

	return &setu_utils.Image{Images: img}, nil
}

// 检查响应是否成功
func (r *Response) IsSuccess() bool {
	return r.Code == Success
}

// 返回响应对应的错误
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

// 获取图片
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
