package setu_utils

import (
	"net/http"
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
