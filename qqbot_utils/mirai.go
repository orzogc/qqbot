package qqbot_utils

import (
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

func SendPrivateText(qqClient *client.QQClient, qq int64, text string) {
	logger := logger.WithField("from", "SendPrivateMessage")
	reply := message.NewSendingMessage()
	reply.Append(message.NewText(text))
	logger.Infof("给QQ %d 发送消息 %s", qq, text)
	if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
		logger.Errorf("给QQ %d 发送消息失败", qq)
	}
}

func SendGroupText(qqClient *client.QQClient, msg *message.GroupMessage, text string) {
	logger := logger.WithField("from", "SendGroupText")
	reply := message.NewSendingMessage()
	reply.Append(message.NewReply(msg))
	reply.Append(message.NewAt(msg.Sender.Uin, "@"+msg.Sender.DisplayName()))
	reply.Append(message.NewText(text))
	logger.Infof("给QQ群 %d 里的QQ %d 发送消息 %s", msg.GroupCode, msg.Sender.Uin, text)
	if result := qqClient.SendGroupMessage(msg.GroupCode, reply); result == nil || result.Id <= 0 {
		logger.Errorf("给QQ群 %d 发送消息失败", msg.GroupCode)
	}
}
