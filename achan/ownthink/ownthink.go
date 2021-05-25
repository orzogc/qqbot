package ownthink

import (
	"encoding/json"
	"fmt"

	"github.com/orzogc/qqbot/qqbot_utils"
)

const OwnthinkURL = "https://api.ownthink.com/bot" // 思知机器人接口

// 思知机器人
type Ownthink struct {
	Spoken string `json:"spoken"` // 请求的文本
	AppID  string `json:"appid"`  // 机器人的appid，填写可使用自己的机器人
	UserID string `json:"userid"` // 自己管理的用户id，填写可进行上下文对话
}

// 返回信息
type Info struct {
	Text string `json:"text"` // 返回的文本信息
}

// 返回数据
type Data struct {
	Type int  `json:"type"` // 返回的数据类型，5000表示正确返回文本类型的答复
	Info Info `json:"info"` // 返回信息
}

// 思知机器人的响应
type Response struct {
	Message string `json:"message"` // success表示请求正确，error表示请求错误
	Data    Data   `json:"data"`    // 返回数据
}

// 请求
func (o *Ownthink) Request() (*Response, error) {
	body, err := qqbot_utils.PostJSON(OwnthinkURL, o)
	if err != nil {
		return nil, err
	}

	resp := new(Response)
	err = json.Unmarshal(body, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// 聊天，实现AI接口
func (o *Ownthink) Chat() (string, error) {
	resp, err := o.Request()
	if err != nil {
		return "", err
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("思知机器人接口返回错误：%s", resp.Message)
	}

	return resp.Data.Info.Text, nil
}

// 聊天，实现AI接口
func (o *Ownthink) ChatWith(text string, id string) (string, error) {
	ownthink := *o
	ownthink.Spoken = text
	ownthink.UserID = id

	return ownthink.Chat()
}

// 检查响应是否成功
func (r *Response) IsSuccess() bool {
	return r.Message == "success"
}
