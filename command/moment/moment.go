package moment

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/moment/moment_utils"
	"github.com/orzogc/qqbot/command/moment/square"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const MomentID = "moment" // ID

var logger = utils.GetModuleLogger(MomentID) // 日志记录

var _ Moment = (*square.AcFunSquare)(nil)

// 动态接口
type Moment interface {
	// 获取动态
	GetMoment() ([]moment_utils.Moment, error)
}

// 回复配置
type Reply struct {
	GetMomentFailed  string `json:"getMomentFailed"`  // 获取动态失败的回复
	SendMomentFailed string `json:"sendMomentFailed"` // 发送动态失败的回复
}

// 配置
type Config struct {
	Commands map[string][]string `json:"commands"` // 命令关键字
	Reply    Reply               `json:"reply"`    // 回复配置
}

// 动态机器人
type MomentBot struct {
	config *Config // 配置
}

// 新建动态机器人
func NewMomentBot(config *Config) *MomentBot {
	return &MomentBot{
		config: config,
	}
}

// 部分配置没有设置的话采用默认配置，设置cmd，返回设置好的cmd，实现Command接口
func (b *MomentBot) SetConfig(cmd map[string][]interface{}) map[string][]interface{} {
	if len(b.config.Commands) == 0 {
		b.config.Commands = map[string][]string{
			square.AcFunSquareID: {"广场", "square"},
		}
	}
	if b.config.Reply.GetMomentFailed == "" {
		b.config.Reply.GetMomentFailed = "获取动态失败"
	}
	if b.config.Reply.SendMomentFailed == "" {
		b.config.Reply.SendMomentFailed = "发送动态失败"
	}

	momentCmd := map[string]Moment{
		square.AcFunSquareID: &square.AcFunSquare{},
	}
	for k, v := range b.config.Commands {
		command, ok := momentCmd[k]
		if !ok {
			logger.Warnf("未知的动态机器人命令ID：%s", k)
			continue
		}
		for _, s := range v {
			if c, ok := cmd[s]; ok {
				c = append(c, command)
				cmd[s] = c
			} else {
				cmd[s] = []interface{}{command}
			}
		}
	}

	return cmd
}

// 处理私聊消息，实现Command接口
func (b *MomentBot) HandlePrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, cmd map[interface{}]struct{}, keyword string) {
	logger := logger.WithField("from", "HandlePrivateMessage")

	momentCmd := make(map[Moment]struct{})
	for c := range cmd {
		if c, ok := c.(Moment); ok {
			momentCmd[c] = struct{}{}
		}
	}
	if len(momentCmd) == 0 {
		return
	}

	moments, err := getMoment(momentCmd, keyword)
	if err != nil {
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("获取动态失败")
		qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.GetMomentFailed)
		if len(moments) == 0 {
			return
		}
	}
	if len(moments) != 0 {
		b.sendPrivateMessage(qqClient, msg, moments)
	}
}

// 处理群聊消息，实现Command接口
func (b *MomentBot) HandleGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, cmd map[interface{}]struct{}, keyword string) {
	logger := logger.WithField("from", "HandleGroupMessage")

	momentCmd := make(map[Moment]struct{})
	for c := range cmd {
		if c, ok := c.(Moment); ok {
			momentCmd[c] = struct{}{}
		}
	}
	if len(momentCmd) == 0 {
		return
	}

	moments, err := getMoment(momentCmd, keyword)
	if err != nil {
		logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("获取动态失败")
		qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.GetMomentFailed)
		if len(moments) == 0 {
			return
		}
	}
	if len(moments) != 0 {
		b.sendGroupMessage(qqClient, msg, moments)
	}
}

// 发送私聊信息
func (b *MomentBot) sendPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, moments []moment_utils.Moment) {
	logger := logger.WithField("from", "sendPrivateMessage")

	reply := message.NewSendingMessage()
	for _, m := range moments {
		reply.Append(message.NewText(m.ToString()))
		for _, img := range m.Images {
			if len(img) != 0 {
				element, err := qqClient.UploadPrivateImage(msg.Sender.Uin, bytes.NewReader(img))
				if err != nil {
					logger.WithError(err).WithField("receiverQQ", msg.Sender.Uin).Error("上传私聊图片失败")
					continue
				}
				reply.Append(element)
			}
		}
	}
	logger.WithField("receiverQQ", msg.Sender.Uin).Info("发送动态")
	if result := qqClient.SendPrivateMessage(msg.Sender.Uin, reply); result == nil || result.Id <= 0 {
		logger.WithField("receiverQQ", msg.Sender.Uin).Error("发送动态失败")
		qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.SendMomentFailed)
	}
}

// 发送群聊消息
func (b *MomentBot) sendGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, moments []moment_utils.Moment) {
	logger := logger.WithField("from", "sendGroupMessage")

	reply := message.NewSendingMessage()
	reply.Append(message.NewReply(msg))
	for _, m := range moments {
		reply.Append(message.NewText(m.ToString()))
		for _, img := range m.Images {
			if len(img) != 0 {
				element, err := qqClient.UploadGroupImage(msg.GroupCode, bytes.NewReader(img))
				if err != nil {
					logger.WithError(err).
						WithField("qqGroup", msg.GroupCode).
						WithField("receiverQQ", msg.Sender.Uin).
						Error("上传群聊图片失败")
					continue
				}
				reply.Append(element)
			}
		}
	}
	logger.WithField("qqGroup", msg.GroupCode).
		WithField("receiverQQ", msg.Sender.Uin).
		Info("发送动态")
	if result := qqClient.SendGroupMessage(msg.GroupCode, reply); result == nil || result.Id <= 0 {
		logger.
			WithField("qqGroup", msg.GroupCode).
			WithField("receiverQQ", msg.Sender.Uin).
			Error("发送动态失败")
		qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.SendMomentFailed)
	}
}

// 获取动态
func getMoment(cmd map[Moment]struct{}, keyword string) (moments []moment_utils.Moment, err error) {
	logger := logger.WithField("from", "getMoment")

	var e error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for c := range cmd {
		wg.Add(1)
		go func(c Moment) {
			defer wg.Done()
			m, err := c.GetMoment()
			if err != nil {
				logger.WithError(err).Error("获取动态失败")
				mu.Lock()
				defer mu.Unlock()
				e = err
				return
			}
			mu.Lock()
			defer mu.Unlock()
			moments = append(moments, m...)
		}(c)
	}
	wg.Wait()

	if len(moments) == 0 {
		if e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("获取动态失败")
	}

	return moments, nil
}
