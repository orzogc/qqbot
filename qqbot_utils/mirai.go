package qqbot_utils

import (
	"strings"

	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// 发送私聊文字
func SendPrivateText(qqClient *client.QQClient, qq int64, text string) bool {
	logger := logger.WithField("from", "SendPrivateText")
	reply := message.NewSendingMessage()
	reply.Append(message.NewText(text))
	logger.WithField("receiverQQ", qq).WithField("text", text).Info("发送私聊消息")
	if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
		logger.WithField("receiverQQ", qq).WithField("text", text).Error("发送私聊消息失败")
		return false
	}

	return true
}

// 回复群聊文字
func ReplyGroupText(qqClient *client.QQClient, msg *message.GroupMessage, text string) bool {
	logger := logger.WithField("from", "ReplyGroupText")
	reply := message.NewSendingMessage()
	reply.Append(message.NewReply(msg))
	reply.Append(message.NewText(text))
	logger.WithField("qqGroup", msg.GroupCode).
		WithField("receiverQQ", msg.Sender.Uin).
		WithField("text", text).
		Info("发送群聊消息")
	if result := qqClient.SendGroupMessage(msg.GroupCode, reply); result == nil || result.Id <= 0 {
		logger.WithField("qqGroup", msg.GroupCode).
			WithField("receiverQQ", msg.Sender.Uin).
			WithField("text", text).
			Error("发送群聊消息失败")
		return false
	}

	return true
}

// 获取私聊消息里的文本
func GetPrivateText(msg *message.PrivateMessage) string {
	var texts []string
	for _, element := range msg.Elements {
		if e, ok := element.(*message.TextElement); ok {
			texts = append(texts, e.Content)
		}
	}

	return strings.Join(texts, " ")
}

// 获取群聊消息里的文本
func GetGroupText(msg *message.GroupMessage) string {
	var texts []string
	for _, element := range msg.Elements {
		if e, ok := element.(*message.TextElement); ok {
			texts = append(texts, e.Content)
		}
	}

	return strings.Join(texts, " ")
}

// 获取群聊消息里@指定qq的文本
func GetGroupAtText(qq int64, msg *message.GroupMessage) (text string, isAt bool) {
	var texts []string
	for _, element := range msg.Elements {
		switch e := element.(type) {
		case *message.AtElement:
			if e.Target == qq {
				isAt = true
			}
		case *message.TextElement:
			texts = append(texts, e.Content)
		default:
		}
	}

	if isAt {
		return strings.Join(texts, " "), isAt
	}

	return "", isAt
}
