package qqbot_utils

import "errors"

var (
	AllCommands    = make(map[string]struct{}) // 全部命令
	ErrorNoCommand = errors.New("没有发现有效的命令")
)
