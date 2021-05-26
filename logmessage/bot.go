package logmessage

import (
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const LogMessageID = "logmessage" // ID

var (
	instance = &LogMessageBot{}                    // 机器人实例
	logger   = utils.GetModuleLogger(LogMessageID) // 日志记录
)

// 负责记录消息的bot
type LogMessageBot struct{}

// 初始化
func init() {
	bot.RegisterModule(instance)
}

func (b *LogMessageBot) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       LogMessageID,
		Instance: instance,
	}
}

func (b *LogMessageBot) Init() {}

func (b *LogMessageBot) PostInit() {}

func (b *LogMessageBot) Serve(bot *bot.Bot) {
	registerBot(bot)
}

func (b *LogMessageBot) Start(bot *bot.Bot) {}

func (b *LogMessageBot) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

// 处理私聊
func onPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage) {
	logger := logger.WithField("from", "onPrivateMessage")
	text := qqbot_utils.GetPrivateText(msg)
	logger.WithField("senderQQ", msg.Sender.Uin).
		WithField("text", text).
		Infof("接收私聊消息：%s", msg.ToString())
}

// 处理群聊
func onGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage) {
	logger := logger.WithField("from", "onGroupMessage")
	if text, isAt := qqbot_utils.GetGroupAtText(qqClient.Uin, msg); isAt {
		logger.WithField("qqGroup", msg.GroupCode).
			WithField("senderQQ", msg.Sender.Uin).
			WithField("text", text).
			Infof("接收群聊消息：%s", msg.ToString())
	}
}

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
