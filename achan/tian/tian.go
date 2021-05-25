package tian

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/orzogc/qqbot/qqbot_utils"
)

const (
	TianURL  = "https://api.tianapi.com/txapi/robot/index" // 天行机器人接口
	key      = "key"
	question = "question"
	uniqueid = "uniqueid"
	mode     = "mode"
	priv     = "priv"
	restype  = "restype"
)

// 天行机器人
type Tian struct {
	Key      string `json:"key"`      // API密钥，需自行注册获得
	Question string `json:"question"` // 聊天文本内容
	UniqueID string `json:"uniqueid"` // 用户唯一身份ID，方便上下文关联
	Mode     uint   `json:"mode"`     // 工作模式，宽松模式0（回答率高）、精确模式1（相关性高）、私有模式2（只从私有词库中回答）
	Priv     uint   `json:"priv"`     // 私有词库匹配模式，完整匹配0[默认]、智能匹配1，模糊匹配2，结尾匹配3，开头匹配4
	Restype  uint   `json:"restype"`  // 输入类型，文本0、语音1、人脸图片2
}

// 天行机器人的响应
type Response struct {
	Code     Code       `json:"code"`     // 返回码
	Msg      string     `json:"msg"`      // 返回信息
	NewsList []NewsList `json:"newslist"` // 天行机器人的回复
}

// 天行机器人的回复
type NewsList struct {
	Reply    string `json:"reply"`    // 回复内容
	Datatype string `json:"datatype"` // 回复内容类型，text文本；view图文；image图片
}

// 返回码
type Code int

// 返回成功
const Success Code = 200

func (t *Tian) query() (url.Values, error) {
	query := url.Values{}
	if t.Key == "" {
		return nil, fmt.Errorf("Key不能为空")
	}
	query.Add(key, t.Key)
	if t.Question == "" {
		return nil, fmt.Errorf("Question不能为空")
	}
	query.Add(question, t.Question)
	if t.UniqueID != "" {
		query.Add(uniqueid, t.UniqueID)
	}
	if t.Mode != 0 {
		if t.Mode > 2 {
			return nil, fmt.Errorf("Mode必须为0、1或2，现为%d", t.Mode)
		}
		query.Add(mode, strconv.FormatUint(uint64(t.Mode), 10))
	}
	if t.Priv != 0 {
		if t.Priv > 4 {
			return nil, fmt.Errorf("Priv必须为0到4，现为%d", t.Priv)
		}
		query.Add(priv, strconv.FormatUint(uint64(t.Priv), 10))
	}
	if t.Restype != 0 {
		if t.Restype > 2 {
			return nil, fmt.Errorf("Restype必须为0、1或2，现为%d", t.Restype)
		}
		query.Add(restype, strconv.FormatUint(uint64(t.Restype), 10))
	}

	return query, nil
}

// 请求
func (t *Tian) Request() (*Response, error) {
	query, err := t.query()
	if err != nil {
		return nil, err
	}
	body, err := qqbot_utils.Get(TianURL, query)
	if err != nil {
		return nil, err
	}

	r := new(Response)
	if err = json.Unmarshal(body, r); err != nil {
		return nil, err
	}

	return r, nil
}

// 聊天，实现AI接口
func (t *Tian) Chat() (string, error) {
	resp, err := t.Request()
	if err != nil {
		return "", err
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("天行机器人接口返回错误，错误码：%d，错误信息：%s", resp.Code, resp.Msg)
	}

	replies := make([]string, 0, len(resp.NewsList))
	for _, r := range resp.NewsList {
		replies = append(replies, r.Reply)
	}
	reply := strings.Join(replies, " ")

	return reply, nil
}

// 聊天，实现AI接口
func (t *Tian) ChatWith(text string, id string) (string, error) {
	tian := *t
	tian.Question = text
	tian.UniqueID = id

	return tian.Chat()
}

// 检查响应是否成功
func (r *Response) IsSuccess() bool {
	return r.Code == Success
}
