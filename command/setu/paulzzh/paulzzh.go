package paulzzh

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/orzogc/qqbot/command/setu/setu_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	PaulzzhID  = "paulzzh"                                // ID
	PaulzzhURL = "https://img.paulzzh.tech/touhou/random" // paulzzh接口
	Type       = "type"
	Site       = "site"
	Size       = "size"
	Proxy      = "proxy"
	Tag        = "tag"
)

var ErrorTag = errors.New("paulzzh东方图片搜索关键字包含非英文字母")

// paulzzh的请求query
type Paulzzh struct {
	Type  string `json:"type"`  // 302(默认)，json(支持跨域)
	Site  string `json:"site"`  // konachan仅使用konachan(此方酱)(默认)，yandere仅使用yande.re(病娇酱)，all全部使用
	Size  string `json:"size"`  // pc横屏壁纸(默认)，wap竖屏壁纸，all全部
	Proxy uint   `json:"proxy"` // 1使用反向代理，0的话Type是302返回源站链接
	Tag   string `json:"tag"`   // 无默认值，筛选指定标签的图片，可能较慢，必须是英文
}

// paulzzh请求的响应
type Response struct {
	Author    string `json:"author"`    // 图片作者
	Width     int    `json:"width"`     // 图片宽度
	Height    int    `json:"height"`    // 图片高度
	ID        int    `json:"id"`        // 图片ID
	JpegURL   string `json:"jpegurl"`   // jpeg图片链接
	Md5       string `json:"md5"`       // 图片md5
	Preview   string `json:"preview"`   // 图片预览链接
	Size      int    `json:"size"`      // 图片大小
	Source    string `json:"source"`    // 图片来源
	Tags      string `json:"tags"`      // 图片tags
	Timestamp int64  `json:"timestamp"` // 图片的时间戳
	URL       string `json:"url"`       // 图片链接
}

// 是否英文字母和空格
func isLetter(s string) bool {
	for _, r := range s {
		if (r < 'a' || r > 'z') && (r < 'A' || r > 'Z') && r != ' ' {
			return false
		}
	}
	return true
}

// 获取图片
func (r *Response) GetImage() ([]byte, error) {
	body, err := qqbot_utils.Get(r.URL, nil)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// 设置请求query
func (p *Paulzzh) query() (url.Values, error) {
	query := url.Values{}
	if p.Type != "" {
		if p.Type != "302" && p.Type != "json" {
			return nil, fmt.Errorf("Type必须为302或json，现为%s", p.Type)
		}
		query.Add(Type, p.Type)
	}
	if p.Site != "" {
		if p.Site != "konachan" && p.Site != "yandere" && p.Site != "all" {
			return nil, fmt.Errorf("Site必须为konachan、yandere或all，现为%s", p.Site)
		}
		query.Add(Site, p.Site)
	}
	if p.Size != "" {
		if p.Size != "pc" && p.Size != "wap" && p.Size != "all" {
			return nil, fmt.Errorf("Size必须为pc、wap或all，现为%s", p.Size)
		}
		query.Add(Size, p.Size)
	}
	if p.Proxy != 1 {
		if p.Proxy != 0 {
			return nil, fmt.Errorf("Proxy必须为0或1，现为%d", p.Proxy)
		}
		query.Add(Proxy, "0")
	}
	if p.Tag != "" {
		if !isLetter(p.Tag) {
			return nil, fmt.Errorf("%w：%s", ErrorTag, p.Tag)
		}
		query.Add(Tag, p.Tag)
	}

	return query, nil
}

// 请求
func (p *Paulzzh) Request() (*Response, error) {
	if p.Type != "json" {
		return nil, fmt.Errorf("Type必须是json")
	}
	query, err := p.query()
	if err != nil {
		return nil, err
	}

	body, err := qqbot_utils.Get(PaulzzhURL, query)
	if err != nil {
		return nil, err
	}
	if !json.Valid(body) {
		return nil, fmt.Errorf("请求出现错误")
	}
	resp := new(Response)
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 获取图片
func (p *Paulzzh) GetTouhouImage() ([]byte, error) {
	query, err := p.query()
	if err != nil {
		return nil, err
	}

	if p.Type == "json" {
		body, err := qqbot_utils.Get(PaulzzhURL, query)
		if err != nil {
			return nil, err
		}
		if !json.Valid(body) {
			return body, nil
		}
		resp := new(Response)
		err = json.Unmarshal(body, resp)
		if err != nil {
			return nil, err
		}
		return resp.GetImage()
	}

	body, err := qqbot_utils.Get(PaulzzhURL, query)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// 获取图片，实现Setu接口
func (p *Paulzzh) GetImage(keyword string) (*setu_utils.Image, error) {
	paulzzh := *p
	paulzzh.Tag = keyword
	img, err := paulzzh.GetTouhouImage()
	if err != nil {
		return nil, err
	}

	return &setu_utils.Image{Images: [][]byte{img}}, nil
}
