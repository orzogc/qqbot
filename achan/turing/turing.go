package turing

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"

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

func (r *Request) Chat() (*Response, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, TuringURL, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := qqbot_utils.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	log.Printf("%s", string(body))

	response := new(Response)
	err = json.Unmarshal(body, response)
	if err != nil {
		return nil, err
	}

	return response, nil
}
