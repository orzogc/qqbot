package paulzzh

import (
	"encoding/json"
	"fmt"
	"net/url"

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

type Query struct {
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

func (r *Response) GetImage() ([][]byte, error) {
	body, err := setu_utils.Get(r.URL, nil)
	if err != nil {
		return nil, err
	}

	return [][]byte{body}, nil
}

func (q *Query) GetImage() ([][]byte, error) {
	query := url.Values{}
	if q.Type != "" {
		if q.Type != "302" && q.Type != "json" {
			return nil, fmt.Errorf("Type必须为302或json，现为%s", q.Type)
		}
		query.Add(Type, q.Type)
	}
	if q.Site != "" {
		if q.Site != "konachan" && q.Site != "yandere" && q.Site != "all" {
			return nil, fmt.Errorf("Site必须为konachan、yandere或all，现为%s", q.Site)
		}
		query.Add(Site, q.Site)
	}
	if q.Size != "" {
		if q.Size != "pc" && q.Size != "wap" && q.Size != "all" {
			return nil, fmt.Errorf("Size必须为pc、wap或all，现为%s", q.Size)
		}
		query.Add(Size, q.Size)
	}
	if q.Proxy != 1 {
		if q.Proxy != 0 {
			return nil, fmt.Errorf("Proxy必须为0或1，现为%d", q.Proxy)
		}
		query.Add(Proxy, "0")
	}
	if q.Tag != "" {
		query.Add(Tag, q.Tag)
	}

	if q.Type == "json" {
		body, err := setu_utils.Get(PaulzzhURL, query)
		if err != nil {
			return nil, err
		}
		resp := new(Response)
		err = json.Unmarshal(body, resp)
		if err != nil {
			return nil, err
		}
		return resp.GetImage()
	}

	body, err := setu_utils.Get(PaulzzhURL, query)
	if err != nil {
		return nil, err
	}

	return [][]byte{body}, nil
}
