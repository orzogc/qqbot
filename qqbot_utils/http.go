package qqbot_utils

import (
	"io"
	"net/http"
	"net/url"
	"time"
)

var Client = &http.Client{
	Timeout: 20 * time.Second,
}

func SetTimeout(second uint) {
	Client = &http.Client{
		Timeout: time.Duration(second) * time.Second,
	}
}

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
