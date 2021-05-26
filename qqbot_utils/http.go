package qqbot_utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"
)

var (
	UserAgent      = "Mozilla/5.0 (X11; Linux x86_64; rv:88.0) Gecko/20100101 Firefox/88.0" // 默认User-Agent
	Timeout   uint = 20                                                                     // http连接超时秒数
	Client         = &http.Client{Timeout: time.Duration(Timeout) * time.Second}            // http客户端
)

// 设置http客户端的超时
func SetTimeout(second uint) {
	Client = &http.Client{
		Timeout: time.Duration(second) * time.Second,
	}
}

// http get
func Get(url string, query url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if query != nil {
		q := req.URL.Query()
		for k, v := range query {
			for _, s := range v {
				q.Add(k, s)
			}
		}
		req.URL.RawQuery = q.Encode()
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// http post json
func PostJSON(url string, v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// http post form urlencoded
func PostForm(url string, form url.Values) ([]byte, error) {
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader([]byte(form.Encode())))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", UserAgent)

	resp, err := Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
