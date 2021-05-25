package qqbot_utils

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

// 发送私聊文字
func SendPrivateText(qqClient *client.QQClient, qq int64, text string) {
	logger := logger.WithField("from", "SendPrivateText")
	reply := message.NewSendingMessage()
	reply.Append(message.NewText(text))
	logger.WithField("receiverQQ", qq).WithField("text", text).Info("发送私聊消息")
	if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
		logger.WithField("receiverQQ", qq).WithField("text", text).Error("发送私聊消息失败")
	}
}

// 回复群聊文字
func ReplyGroupText(qqClient *client.QQClient, msg *message.GroupMessage, text string) {
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
	}
}
