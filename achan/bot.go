package achan

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/achan/ownthink"
	"github.com/orzogc/qqbot/achan/tian"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/spf13/viper"
)

// ID
const AchanID = "achan"

var (
	instance = &AchanBot{}                    // 机器人实例
	logger   = utils.GetModuleLogger(AchanID) // 日志记录
)

// 配置
type Config struct {
	Tian     tian.Tian         `json:"tian"`     // 天行机器人的配置
	Ownthink ownthink.Ownthink `json:"ownthink"` // 思知机器人的配置
	Replace  map[string]string `json:"replace"`  // 词语替换
	Replace2 map[string]string `json:"replace2"` // 词语替换第二层
}

// 聊天机器人
type AchanBot struct {
	config *Config // 配置
}

// 初始化
func init() {
	bot.RegisterModule(instance)
}

func (a *AchanBot) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       AchanID,
		Instance: instance,
	}
}

func (b *AchanBot) Init() {
	logger := logger.WithField("from", "Init")
	viper := viper.New()
	viper.SetConfigName(AchanID)
	viper.SetConfigType("json")
	path, err := os.Executable()
	if err != nil {
		logger.WithError(err).Panic("获取执行文件所在位置失败")
	}

	dir := filepath.Dir(path)
	viper.AddConfigPath(dir)
	viper.AddConfigPath(".")

	err = viper.ReadInConfig()
	if err != nil {
		logger.WithError(err).Warn("读取设置文件setu.json失败，使用默认设置")
		instance.config = new(Config)
		instance.config.Replace = make(map[string]string)
		instance.config.Replace2 = make(map[string]string)
	} else {
		err = viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件setu.json的内容无效，使用默认设置")
			instance.config = new(Config)
			instance.config.Replace = make(map[string]string)
			instance.config.Replace2 = make(map[string]string)
		}
	}
}

func (b *AchanBot) PostInit() {}

func (b *AchanBot) Serve(bot *bot.Bot) {
	registerBot(bot)
}

func (b *AchanBot) Start(bot *bot.Bot) {}

func (b *AchanBot) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

// 替换词语
func replace(s string) string {
	for k, v := range instance.config.Replace {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

// 替换词语
func replace2(s string) string {
	for k, v := range instance.config.Replace2 {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

// 处理私聊
func onPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage) {
	logger := logger.WithField("from", "onPrivateMessage")

	text := qqbot_utils.GetPrivateText(msg)
	if strings.Contains(text, "#") {
		return
	}

	reply, err := instance.config.Ownthink.ChatWith(text, strconv.FormatInt(msg.Sender.Uin, 10))
	if err != nil {
		logger.WithError(err).Error("请求出现错误")
		return
	}
	reply = replace(reply)
	reply = replace2(reply)

	qqbot_utils.SendPrivateText(qqClient, msg, reply)
}

// 处理群聊
func onGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage) {
	logger := logger.WithField("from", "onGroupMessage")

	if text, isAt := qqbot_utils.GetGroupAtText(qqClient.Uin, msg); isAt {
		if strings.Contains(text, "#") {
			return
		}

		reply, err := instance.config.Ownthink.ChatWith(text, strconv.FormatInt(msg.Sender.Uin, 10))
		if err != nil {
			logger.WithError(err).Error("请求出现错误")
			return
		}
		reply = replace(reply)
		reply = replace2(reply)

		qqbot_utils.ReplyGroupText(qqClient, msg, reply)
	}
}

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
