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
	"github.com/orzogc/qqbot/achan/tian"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/spf13/viper"
)

const AchanID = "achan"

var (
	instance = &AchanBot{}
	logger   = utils.GetModuleLogger(AchanID)
)

type Config struct {
	Tian    tian.Query        `json:"tian"`
	Replace map[string]string `json:"replace"`
}

type AchanBot struct {
	config *Config
}

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
		instance.config.Tian = tian.Query{}
	} else {
		err = viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件setu.json的内容无效，使用默认设置")
			instance.config = new(Config)
			instance.config.Tian = tian.Query{}
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

func replace(s string) string {
	for k, v := range instance.config.Replace {
		s = strings.ReplaceAll(s, k, v)
	}

	return s
}

func onPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage) {
	logger := logger.WithField("from", "onPrivateMessage")

	var texts []string
	for _, element := range msg.Elements {
		if e, ok := element.(*message.TextElement); ok {
			texts = append(texts, e.Content)
		}
	}
	text := strings.Join(texts, " ")
	if strings.Contains(text, "#") {
		return
	}

	query := instance.config.Tian
	query.Question = text
	query.UniqueID = strconv.FormatInt(msg.Sender.Uin, 10)
	reply, err := query.Chat()
	if err != nil {
		logger.WithError(err).Error("请求出现错误")
		return
	}
	reply = replace(reply)

	qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, reply)
}

func onGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage) {
	logger := logger.WithField("from", "onGroupMessage")

	var isAt bool
	var texts []string
	for _, element := range msg.Elements {
		switch e := element.(type) {
		case *message.AtElement:
			if e.Target == qqClient.Uin {
				isAt = true
			}
		case *message.TextElement:
			texts = append(texts, e.Content)
		default:
		}
	}

	if isAt {
		text := strings.Join(texts, " ")
		if strings.Contains(text, "#") {
			return
		}

		query := instance.config.Tian
		query.Question = text
		query.UniqueID = strconv.FormatInt(msg.Sender.Uin, 10)
		reply, err := query.Chat()
		if err != nil {
			logger.WithError(err).Error("请求出现错误")
			return
		}
		reply = replace(reply)

		qqbot_utils.SendGroupText(qqClient, msg, reply)
	}
}

func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
