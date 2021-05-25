package achan

import (
	"github.com/orzogc/qqbot/achan/ownthink"
	"github.com/orzogc/qqbot/achan/tian"
	"github.com/orzogc/qqbot/achan/turing"
)

var (
	_ AI = (*tian.Tian)(nil)
	_ AI = (*turing.Turing)(nil)
	_ AI = (*ownthink.Ownthink)(nil)
)

// AI接口
type AI interface {
	// 聊天，返回AI的回答文本
	Chat() (string, error)
	// 聊天，text为对话文本，id为聊天的ID，返回AI的回答文本
	ChatWith(text string, id string) (string, error)
}
