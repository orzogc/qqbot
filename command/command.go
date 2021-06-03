package command

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/moment"
	"github.com/orzogc/qqbot/command/search"
	"github.com/orzogc/qqbot/command/setu"
	"github.com/orzogc/qqbot/command/zuan"
)

var (
	_ Command = (*setu.SetuBot)(nil)
	_ Command = (*search.SearchBot)(nil)
	_ Command = (*moment.MomentBot)(nil)
	_ Command = (*zuan.ZuanBot)(nil)
)

// Command接口
type Command interface {
	// 部分配置没有设置的话采用默认配置，设置cmd，返回设置好的cmd
	SetConfig(cmd map[string][]interface{}) map[string][]interface{}
	// 处理私聊消息
	HandlePrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, cmd map[interface{}]struct{}, keyword string)
	// 处理群聊消息
	HandleGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, cmd map[interface{}]struct{}, keyword string)
}
