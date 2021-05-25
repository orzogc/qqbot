package reconnect

import (
	"sync"
	"time"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
)

const (
	ReconnectID = "reconnect" // ID
	waitTime    = 10 * time.Second
)

var (
	instance = &ReconnectBot{}
	logger   = utils.GetModuleLogger(ReconnectID)
)

// 负责重连帐号的bot
type ReconnectBot struct{}

// 初始化
func init() {
	bot.RegisterModule(instance)
}

func (b *ReconnectBot) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       ReconnectID,
		Instance: instance,
	}
}

func (b *ReconnectBot) Init() {}

func (b *ReconnectBot) PostInit() {}

func (b *ReconnectBot) Serve(bot *bot.Bot) {
	registerBot(bot)
}

func (b *ReconnectBot) Start(bot *bot.Bot) {}

func (b *ReconnectBot) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

func onDisconnected(qqClient *client.QQClient, event *client.ClientDisconnectedEvent) {
	logger := logger.WithField("from", "onDisconnected")
	logger.WithField("reason", event.Message).Warn("bot已离线，尝试重连")
	time.Sleep(waitTime)

	resp, err := qqClient.Login()
	if err != nil {
		logger.Error("bot重连失败，请重启本bot")
		return
	}
	if !resp.Success {
		switch resp.Error {
		case client.NeedCaptcha:
			logger.Error("bot重连失败：需要验证码，请重启本bot")
		case client.UnsafeDeviceError:
			logger.Error("bot重连失败：设备锁")
			logger.Errorf("bot的QQ帐号已开启设备锁，请前往 %s 验证并重启本bot", resp.VerifyUrl)
		default:
			logger.Errorf("bot重连失败，请重启本bot，响应为：%+v", resp)
		}
	} else {
		logger.Info("bot重连成功")
	}
}

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnDisconnected(onDisconnected)
}
