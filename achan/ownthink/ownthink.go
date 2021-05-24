package ownthink

import (
	"encoding/json"
	"fmt"

	"github.com/orzogc/qqbot/qqbot_utils"
)

const OwnthinkURL = "https://api.ownthink.com/bot"

type Request struct {
	Spoken string `json:"spoken"`
	AppID  string `json:"appid"`
	UserID string `json:"userid"`
}

type Info struct {
	Text string `json:"text"`
}

type Data struct {
	Type int  `json:"type"`
	Info Info `json:"info"`
}

type Response struct {
	Message string `json:"message"`
	Data    Data   `json:"data"`
}

func (r *Response) IsSuccess() bool {
	return r.Message == "success"
}

func (r *Request) Chat() (string, error) {
	body, err := qqbot_utils.PostJSON(OwnthinkURL, r)
	if err != nil {
		return "", err
	}

	resp := new(Response)
	err = json.Unmarshal(body, resp)
	if err != nil {
		return "", err
	}
	if !resp.IsSuccess() {
		return "", fmt.Errorf("思知机器人接口返回错误：%s", resp.Message)
	}

	return resp.Data.Info.Text, nil
}
