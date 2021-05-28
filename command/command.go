package command

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/moment"
	"github.com/orzogc/qqbot/command/search"
	"github.com/orzogc/qqbot/command/setu"
)

var (
	_ Command = (*setu.SetuBot)(nil)
	_ Command = (*search.SearchBot)(nil)
	_ Command = (*moment.MomentBot)(nil)
)

// Command接口
type Command interface {
	// 部分配置没有设置的话采用默认配置
	SetConfig()
	// 处理私聊消息
	HandlePrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, cmd map[interface{}]struct{}, keyword string)
	// 处理群聊消息
	HandleGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, cmd map[interface{}]struct{}, keyword string)
}
