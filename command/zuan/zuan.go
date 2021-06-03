package zuan

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/zuan/zuanbot"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const ZuanID = "zuan" // ID

var logger = utils.GetModuleLogger(ZuanID) // 日志记录

var _ Zuan = (*zuanbot.Zuanbot)(nil)

// 祖安接口
type Zuan interface {
	// 获取祖安语句
	GetText() (string, error)
}

// 回复配置
type Reply struct {
	GetTextFailed    string `json:"getTextFailed"`    // 获取祖安语句失败的回复
	SendTextFailed   string `json:"sendTextFailed"`   // 发送祖安语句失败的回复
	FindObjectFailed string `json:"findObjectFailed"` // 寻找祖安对象失败的回复
}

// 配置
type Config struct {
	Commands map[string][]string `json:"commands"` // 命令关键字
	Object   []int64             `json:"object"`   // 祖安对象QQ号（仅对QQ群有效）
	Replace  map[string]string   `json:"replace"`  // 词语替换
	Replace2 map[string]string   `json:"replace2"` // 词语替换第二层
	Reply    Reply               `json:"reply"`    // 回复配置
}

// 祖安机器人
type ZuanBot struct {
	config *Config // 配置
}

// 新建搜索机器人
func NewZuanBot(config *Config) *ZuanBot {
	return &ZuanBot{
		config: config,
	}
}

// 部分配置没有设置的话采用默认配置，设置cmd，返回设置好的cmd，实现Command接口
func (b *ZuanBot) SetConfig(cmd map[string][]interface{}) map[string][]interface{} {
	if len(b.config.Commands) == 0 {
		b.config.Commands = map[string][]string{
			zuanbot.ZuanbotID: {"祖安", "霸凌", "凌辱", "80"},
		}
	}
	if b.config.Reply.GetTextFailed == "" {
		b.config.Reply.GetTextFailed = "获取祖安语句失败"
	}
	if b.config.Reply.SendTextFailed == "" {
		b.config.Reply.SendTextFailed = "发送祖安语句失败"
	}
	if b.config.Reply.FindObjectFailed == "" {
		b.config.Reply.FindObjectFailed = "寻找祖安对象失败"
	}

	zuanCmd := map[string]Zuan{
		zuanbot.ZuanbotID: &zuanbot.Zuanbot{},
	}
	for k, v := range b.config.Commands {
		command, ok := zuanCmd[k]
		if !ok {
			logger.Warnf("未知的祖安机器人命令ID：%s", k)
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
func (b *ZuanBot) HandlePrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, cmd map[interface{}]struct{}, keyword string) {
	logger := logger.WithField("from", "HandlePrivateMessage")

	zuanCmd := make(map[Zuan]struct{})
	for c := range cmd {
		if c, ok := c.(Zuan); ok {
			zuanCmd[c] = struct{}{}
		}
	}
	if len(zuanCmd) == 0 {
		return
	}

	text, err := getText(zuanCmd)
	if err != nil {
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("获取祖安语句失败")
		qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.GetTextFailed)
		if text == "" {
			return
		}
	}
	if text != "" {
		text = b.replace(text)
		text = b.replace2(text)
		if ok := qqbot_utils.SendPrivateText(qqClient, msg, text); !ok {
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.SendTextFailed)
		}
	}
}

// 处理群聊消息，实现Command接口
func (b *ZuanBot) HandleGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, cmd map[interface{}]struct{}, keyword string) {
	logger := logger.WithField("from", "HandleGroupMessage")

	zuanCmd := make(map[Zuan]struct{})
	for c := range cmd {
		if c, ok := c.(Zuan); ok {
			zuanCmd[c] = struct{}{}
		}
	}
	if len(zuanCmd) == 0 {
		return
	}

	text, err := getText(zuanCmd)
	if err != nil {
		logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("获取祖安语句失败")
		qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.GetTextFailed)
		if text == "" {
			return
		}
	}
	if text != "" {
		text = b.replace(text)
		text = b.replace2(text)
		receiver := make([]int64, 0, len(b.config.Object))
		reply := message.NewSendingMessage()
		for _, o := range b.config.Object {
			info, err := qqClient.GetMemberInfo(msg.GroupCode, o)
			if err != nil {
				continue
			}
			reply.Append(message.NewAt(o, "@"+info.DisplayName()))
			receiver = append(receiver, o)
		}
		if len(receiver) == 0 {
			logger.WithField("qqGroup", msg.GroupCode).
				WithField("receiverQQ", b.config.Object).
				Info("寻找祖安对象失败")
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.FindObjectFailed)
			return
		}
		reply.Append(message.NewText(text))
		logger.WithField("qqGroup", msg.GroupCode).
			WithField("receiverQQ", receiver).
			WithField("text", text).
			Info("发送群聊消息")
		if result := qqClient.SendGroupMessage(msg.GroupCode, reply); result == nil || result.Id <= 0 {
			logger.WithField("qqGroup", msg.GroupCode).
				WithField("receiverQQ", receiver).
				WithField("text", text).
				Error("发送群聊消息失败")
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.SendTextFailed)
		}
	}
}

// 替换词语
func (b *ZuanBot) replace(s string) string {
	for k, v := range b.config.Replace {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

// 替换词语
func (b *ZuanBot) replace2(s string) string {
	for k, v := range b.config.Replace2 {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

func getText(cmd map[Zuan]struct{}) (string, error) {
	logger := logger.WithField("from", "getText")

	texts := make([]string, 0, len(cmd))
	var e error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for c := range cmd {
		wg.Add(1)
		go func(c Zuan) {
			defer wg.Done()
			text, err := c.GetText()
			if err != nil {
				logger.WithError(err).Error("获取祖安语句失败")
				mu.Lock()
				defer mu.Unlock()
				e = err
				return
			}
			mu.Lock()
			defer mu.Unlock()
			texts = append(texts, text)
		}(c)
	}
	wg.Wait()

	if len(texts) == 0 {
		if e != nil {
			return "", e
		}
		return "", fmt.Errorf("获取祖安语句失败")
	}

	return strings.Join(texts, " "), nil
}
