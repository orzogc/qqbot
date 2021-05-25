package paulzzh

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/setu/setu_utils"
)

const (
	ID         = "paulzzh"
	PaulzzhURL = "https://img.paulzzh.tech/touhou/random"
	Type       = "type"
	Site       = "site"
	Size       = "size"
	Proxy      = "proxy"
	Tag        = "tag"
)

var ErrorTag = errors.New("paulzzh东方图片搜索关键字包含非英文字母")

type Paulzzh struct {
	Type  string `json:"type"`
	Site  string `json:"site"`
	Size  string `json:"size"`
	Proxy uint   `json:"proxy"`
	Tag   string `json:"tag"`
}

type Response struct {
	Author    string `json:"author"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	ID        int    `json:"id"`
	JpegURL   string `json:"jpegurl"`
	Md5       string `json:"md5"`
	Preview   string `json:"preview"`
	Size      int    `json:"size"`
	Source    string `json:"source"`
	Tags      string `json:"tags"`
	Timestamp int64  `json:"timestamp"`
	URL       string `json:"url"`
}

func (r *Response) GetImage() ([]byte, error) {
	body, err := qqbot_utils.Get(r.URL, nil)
	if err != nil {
		return nil, err
	}

	return body, nil
}

func (p *Paulzzh) GetTouhouImage() ([]byte, error) {
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
		if !setu_utils.IsLetter(p.Tag) {
			return nil, fmt.Errorf("%w：%s", ErrorTag, p.Tag)
		}
		query.Add(Tag, p.Tag)
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

func (p *Paulzzh) GetImage(keyword string) ([][]byte, error) {
	paulzzh := *p
	paulzzh.Tag = keyword
	img, err := p.GetTouhouImage()
	if err != nil {
		return nil, err
	}

	return [][]byte{img}, nil
}
