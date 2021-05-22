package setu

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/setu/lolicon"
	"github.com/spf13/viper"
)

const SetuID = "setu"

var (
	instance       = &SetuBot{viper: viper.New()}
	logger         = utils.GetModuleLogger(SetuID)
	errorNoCommand = errors.New("没有发现有效的命令")
)

type Reply struct {
	Normal     string `json:"normal"`
	NoCommand  string `json:"noCommand"`
	Unknown    string `json:"unknown"`
	SendFailed string `json:"sendFailed"`
	Error      string `json:"error"`
}

type Config struct {
	Lolicon  *lolicon.Query      `json:"lolicon"`
	Timeout  uint                `json:"timeout"`
	Commands map[string][]string `json:"commands"`
	Reply    Reply               `json:"reply"`
}

type SetuBot struct {
	viper    *viper.Viper
	config   *Config
	commands map[string][]Setu
}

func init() {
	bot.RegisterModule(instance)
}

func (b *SetuBot) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       SetuID,
		Instance: instance,
	}
}

func (b *SetuBot) Init() {
	logger := logger.WithField("from", "Init")
	instance.viper.SetConfigName(SetuID)
	instance.viper.SetConfigType("json")
	path, err := os.Executable()
	if err != nil {
		logger.WithError(err).Panic("获取执行文件所在位置失败")
	}

	dir := filepath.Dir(path)
	instance.viper.AddConfigPath(dir)
	instance.viper.AddConfigPath(".")

	err = instance.viper.ReadInConfig()
	if err != nil {
		logger.WithError(err).Warn("读取设置文件setu.json失败，使用默认设置")
		instance.config = new(Config)
		instance.config.Lolicon = new(lolicon.Query)
	} else {
		err = instance.viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件setu.json的内容无效，使用默认设置")
			instance.config = new(Config)
			instance.config.Lolicon = new(lolicon.Query)
		}
	}

	if instance.config.Timeout != 0 {
		lolicon.SetTimeout(instance.config.Timeout)
	}
	if len(instance.config.Commands) == 0 {
		instance.config.Commands = map[string][]string{lolicon.ID: {"色图", "涩图", "瑟图"}}
	}
	if instance.config.Reply.Normal == "" {
		instance.config.Reply.Normal = "这是您点的图片"
	}
	if instance.config.Reply.NoCommand == "" {
		instance.config.Reply.NoCommand = "未知命令"
	}
	if instance.config.Reply.Unknown == "" {
		instance.config.Reply.Unknown = "发现未知消息类型，目前只支持文本消息"
	}
	if instance.config.Reply.SendFailed == "" {
		instance.config.Reply.SendFailed = "发送图片失败"
	}
	if instance.config.Reply.Error == "" {
		instance.config.Reply.Error = "获取或上传图片失败"
	}

	api := map[string]Setu{
		lolicon.ID: instance.config.Lolicon,
	}
	instance.commands = make(map[string][]Setu)
	for k, v := range instance.config.Commands {
		if a, ok := api[k]; ok {
			for _, s := range v {
				if c, ok := instance.commands[s]; ok {
					c = append(c, a)
					instance.commands[s] = c
				} else {
					instance.commands[s] = []Setu{a}
				}
			}
		}
	}
}

func (b *SetuBot) PostInit() {}

func (b *SetuBot) Serve(bot *bot.Bot) {
	registerBot(bot)
}

func (b *SetuBot) Start(bot *bot.Bot) {}

func (b *SetuBot) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

func onPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage) {
	logger := logger.WithField("from", "onPrivateMessage")
	//logger.Infof("接收私聊信息：%+v", msg)

	var texts []string
	for _, element := range msg.Elements {
		if e, ok := element.(*message.TextElement); ok {
			texts = append(texts, e.Content)
		}
	}
	text := strings.Join(texts, " ")
	texts = strings.Fields(text)

	images, err := getImage(texts)
	if err != nil {
		logger.WithError(err).Error("获取图片失败")
		if errors.Is(err, errorNoCommand) {
			sendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.NoCommand)
		} else {
			sendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.Error)
		}
		return
	}
	sendPrivateImage(qqClient, msg.Sender.Uin, images)
}

func sendPrivateText(qqClient *client.QQClient, qq int64, text string) {
	logger := logger.WithField("from", "sendPrivateMessage")
	reply := message.NewSendingMessage()
	reply.Append(message.NewText(text))
	logger.Infof("给QQ %d 发送消息 %s", qq, text)
	if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
		logger.Errorf("给QQ %d 发送消息失败", qq)
	}
}

func sendPrivateImage(qqClient *client.QQClient, qq int64, images [][]byte) {
	logger := logger.WithField("from", "sendPrivateImage")
	reply := message.NewSendingMessage()
	reply.Append(message.NewText(instance.config.Reply.Normal))
	for _, img := range images {
		r := bytes.NewReader(img)
		element, err := qqClient.UploadPrivateImage(qq, r)
		if err != nil {
			logger.WithError(err).Error("上传私聊图片失败")
			sendPrivateText(qqClient, qq, instance.config.Reply.Error)
			return
		}
		reply.Append(element)
	}
	logger.Infof("给QQ %d 发送 %d 张图片", qq, len(images))
	if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
		logger.Errorf("给QQ %d 发送图片失败", qq)
		sendPrivateText(qqClient, qq, instance.config.Reply.SendFailed)
	}
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
		//logger.Infof("接收群聊信息：%+v", msg)
		text := strings.Join(texts, " ")
		texts = strings.Fields(text)
		images, err := getImage(texts)
		if err != nil {
			logger.WithError(err).Error("获取图片失败")
			if errors.Is(err, errorNoCommand) {
				sendGroupText(qqClient, msg.GroupCode, msg.Sender.Uin, msg.Sender.DisplayName(), instance.config.Reply.NoCommand)
			} else {
				sendGroupText(qqClient, msg.GroupCode, msg.Sender.Uin, msg.Sender.DisplayName(), instance.config.Reply.Error)
			}
			return
		}
		sendGroupImage(qqClient, msg.GroupCode, msg.Sender.Uin, msg.Sender.DisplayName(), images)
	}
}

func sendGroupText(qqClient *client.QQClient, qqGroup int64, qq int64, qqName string, text string) {
	logger := logger.WithField("from", "sendGroupText")
	reply := message.NewSendingMessage()
	reply.Append(message.NewAt(qq, "@"+qqName))
	reply.Append(message.NewText(text))
	logger.Infof("给QQ群 %d 里的QQ %d 发送消息 %s", qqGroup, qq, text)
	if result := qqClient.SendGroupMessage(qqGroup, reply); result == nil || result.Id <= 0 {
		logger.Errorf("给QQ群 %d 发送消息失败", qqGroup)
	}
}

func sendGroupImage(qqClient *client.QQClient, qqGroup int64, qq int64, qqName string, images [][]byte) {
	logger := logger.WithField("from", "sendGroupImage")
	reply := message.NewSendingMessage()
	reply.Append(message.NewAt(qq, "@"+qqName))
	reply.Append(message.NewText(instance.config.Reply.Normal))
	for _, img := range images {
		r := bytes.NewReader(img)
		element, err := qqClient.UploadGroupImage(qqGroup, r)
		if err != nil {
			logger.WithError(err).Error("上传群聊图片失败")
			sendGroupText(qqClient, qqGroup, qq, qqName, instance.config.Reply.Error)
			return
		}
		reply.Append(element)
	}
	logger.Infof("给QQ群 %d 里的QQ %d 发送 %d 张图片", qqGroup, qq, len(images))
	if result := qqClient.SendGroupMessage(qqGroup, reply); result == nil || result.Id <= 0 {
		logger.Errorf("给QQ群 %d 发送图片失败", qqGroup)
		sendGroupText(qqClient, qqGroup, qq, qqName, instance.config.Reply.SendFailed)
	}
}

func getImage(texts []string) ([][]byte, error) {
	logger := logger.WithField("from", "getImage")

	var hasCommand bool
	keywords := make([]string, 0, len(texts))
	cmd := make(map[Setu]struct{})
	for _, t := range texts {
		var isCommand bool
		for k, v := range instance.commands {
			if strings.Contains(t, k) {
				hasCommand = true
				isCommand = true
				for _, c := range v {
					cmd[c] = struct{}{}
				}
			}
		}
		if !isCommand {
			keywords = append(keywords, t)
		}
	}
	if !hasCommand {
		return nil, errorNoCommand
	}

	var images [][]byte
	for k := range cmd {
		switch c := k.(type) {
		case *lolicon.Query:
			if c.Keyword != "" {
				keywords = append(keywords, c.Keyword)
				c.Keyword = strings.Join(keywords, " ")
			}
			img, err := c.GetImage()
			if err != nil {
				logger.WithError(err).Error("获取图片失败")
				break
			}
			images = append(images, img...)
		default:
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("获取图片失败")
	}

	return images, nil
}

func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
