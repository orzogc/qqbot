package turing

import (
	"encoding/json"
	"strings"

	"github.com/orzogc/qqbot/qqbot_utils"
)

const TuringURL = "https://api.turingos.cn/turingos/api/v2"

type Request struct {
	Data      Data   `json:"data"`
	Key       string `json:"key"`
	Timestamp string `json:"timestamp"`
}

type Data struct {
	Content  []Content `json:"content"`
	UserInfo UserInfo  `json:"userInfo"`
}

type Content struct {
	Data string `json:"data"`
}

type UserInfo struct {
	UniqueID string `json:"uniqueId"`
}

type Response struct {
	GlobalID string   `json:"globalId"`
	Intent   Intent   `json:"intent"`
	Results  []Result `json:"results"`
}

type Intent struct {
	Code         int `json:"code"`
	OperateState int `json:"operateState"`
}

type Result struct {
	GroupType  int    `json:"groupType"`
	Values     Value  `json:"values"`
	ResultType string `json:"resultType"`
}

type Value struct {
	Text      string `json:"text"`
	EmotionID int    `json:"emotionId"`
}

func (r *Request) Chat() (string, error) {
	body, err := qqbot_utils.PostJSON(TuringURL, r)
	if err != nil {
		return "", err
	}

	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return "", err
	}

	replies := make([]string, 0, len(response.Results))
	for _, r := range response.Results {
		replies = append(replies, r.Values.Text)
	}
	reply := strings.Join(replies, " ")

	return reply, nil
}
