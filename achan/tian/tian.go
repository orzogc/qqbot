package tian

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	TianURL  = "http://api.tianapi.com/txapi/robot/index"
	Key      = "key"
	Question = "question"
	UniqueID = "uniqueid"
	Mode     = "mode"
	Priv     = "priv"
	Restype  = "restype"
)

type Query struct {
	Key      string `json:"key"`
	Question string `json:"question"`
	UniqueID string `json:"uniqueid"`
	Mode     uint   `json:"mode"`
	Priv     uint   `json:"priv"`
	Restype  uint   `json:"restype"`
}

type Response struct {
	Code     Code       `json:"code"`
	Msg      string     `json:"msg"`
	NewsList []NewsList `json:"newslist"`
}

type NewsList struct {
	Reply    string `json:"reply"`
	Datatype string `json:"datatype"`
}

type Code int

const Success Code = 200

func (q *Query) Chat() (*Response, error) {
	query := url.Values{}
	if q.Key == "" {
		return nil, fmt.Errorf("Key不能为空")
	}
	query.Add(Key, q.Key)
	if q.Question == "" {
		return nil, fmt.Errorf("Question不能为空")
	}
	query.Add(Question, q.Question)
	if q.UniqueID != "" {
		query.Add(UniqueID, q.UniqueID)
	}
	if q.Mode != 0 {
		if q.Mode > 2 {
			return nil, fmt.Errorf("Mode必须为0、1或2，现为%d", q.Mode)
		}
		query.Add(Mode, strconv.FormatUint(uint64(q.Mode), 10))
	}
	if q.Priv != 0 {
		if q.Priv > 4 {
			return nil, fmt.Errorf("Priv必须为0到4，现为%d", q.Priv)
		}
		query.Add(Priv, strconv.FormatUint(uint64(q.Priv), 10))
	}
	if q.Restype != 0 {
		if q.Restype > 2 {
			return nil, fmt.Errorf("Restype必须为0、1或2，现为%d", q.Restype)
		}
		query.Add(Restype, strconv.FormatUint(uint64(q.Restype), 10))
	}

	body, err := qqbot_utils.Get(TianURL, query)
	if err != nil {
		return nil, err
	}

	r := new(Response)
	if err = json.Unmarshal(body, r); err != nil {
		return nil, err
	}
	if !r.IsSuccess() {
		return nil, fmt.Errorf("天行机器人接口返回错误，错误码：%d，错误信息：%s", r.Code, r.Msg)
	}

	return r, nil
}

func (r *Response) IsSuccess() bool {
	return r.Code == Success
}
