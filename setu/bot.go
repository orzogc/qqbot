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
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/setu/islandwind233"
	"github.com/orzogc/qqbot/setu/lolicon"
	"github.com/orzogc/qqbot/setu/paulzzh"
	"github.com/orzogc/qqbot/setu/pixiv"
	"github.com/orzogc/qqbot/setu/setu_utils"
	"github.com/spf13/viper"
)

const SetuID = "setu"

var (
	instance = &SetuBot{}
	logger   = utils.GetModuleLogger(SetuID)
)

var (
	errorNoCommand = errors.New("没有发现有效的命令")
	errorTag       = errors.New("东方图片搜索关键字包含非英文字母")
	errorNoTag     = errors.New("pixiv图片搜索没有关键字")
)

type Reply struct {
	Normal            string `json:"normal"`
	NoCommand         string `json:"noCommand"`
	GetImageFailed    string `json:"getImageFailed"`
	UploadImageFailed string `json:"uploadImageFailed"`
	SendImageFailed   string `json:"sendImageFailed"`
	KeywordNotFound   string `json:"keywordNotFound"`
	QuotaLimit        string `json:"quotaLimit"`
	TagError          string `json:"tagError"`
	NoTagError        string `json:"noTagError"`
}

type PixivConfig struct {
	PHPSESSID    string             `json:"phpsessid"`
	SearchOption pixiv.SearchOption `json:"searchOption"`
}

type Config struct {
	Lolicon  lolicon.Query       `json:"lolicon"`
	Paulzzh  paulzzh.Query       `json:"paulzzh"`
	Pixiv    PixivConfig         `json:"pixiv"`
	Commands map[string][]string `json:"commands"`
	Reply    Reply               `json:"reply"`
}

type SetuBot struct {
	config   *Config
	pixiv    *pixiv.Pixiv
	commands map[string][]string
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
	viper := viper.New()
	viper.SetConfigName(SetuID)
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
		instance.config.Lolicon = lolicon.Query{}
	} else {
		err = viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件setu.json的内容无效，使用默认设置")
			instance.config = new(Config)
			instance.config.Lolicon = lolicon.Query{}
		}
	}

	if len(instance.config.Commands) == 0 {
		instance.config.Commands = map[string][]string{
			lolicon.ID:              {"色图", "涩图", "瑟图", "setu"},
			islandwind233.AnimeID:   {"二次元", "二刺猿", "erciyuan"},
			islandwind233.CosplayID: {"cos", "余弦", "三次元"},
			paulzzh.ID:              {"东方", "车万", "東方", "越共", "dongfang", "touhou"},
			pixiv.ID:                {"pixiv", "p站", "P站"},
		}
	}
	if instance.config.Reply.Normal == "" {
		instance.config.Reply.Normal = "这是您点的图片"
	}
	if instance.config.Reply.NoCommand == "" {
		instance.config.Reply.NoCommand = "未知命令"
	}
	if instance.config.Reply.GetImageFailed == "" {
		instance.config.Reply.GetImageFailed = "获取图片失败"
	}
	if instance.config.Reply.UploadImageFailed == "" {
		instance.config.Reply.UploadImageFailed = "上传图片失败"
	}
	if instance.config.Reply.SendImageFailed == "" {
		instance.config.Reply.SendImageFailed = "发送图片失败"
	}
	if instance.config.Reply.KeywordNotFound == "" {
		instance.config.Reply.KeywordNotFound = "找不到符合关键字的图片"
	}
	if instance.config.Reply.QuotaLimit == "" {
		instance.config.Reply.QuotaLimit = "已经达到接口的调用额度上限"
	}
	if instance.config.Reply.TagError == "" {
		instance.config.Reply.TagError = "东方图片搜索关键字必须是英文字母"
	}
	if instance.config.Reply.NoTagError == "" {
		instance.config.Reply.NoTagError = "pixiv图片搜索需要关键字"
	}

	instance.pixiv = pixiv.New(instance.config.Pixiv.PHPSESSID)
	instance.pixiv.SearchOption = &instance.config.Pixiv.SearchOption

	instance.commands = make(map[string][]string)
	for k, v := range instance.config.Commands {
		for _, s := range v {
			if c, ok := instance.commands[s]; ok {
				c = append(c, k)
				instance.commands[s] = c
			} else {
				instance.commands[s] = []string{k}
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

	var texts []string
	for _, element := range msg.Elements {
		if e, ok := element.(*message.TextElement); ok {
			texts = append(texts, e.Content)
		}
	}
	text := strings.Join(texts, " ")
	if !strings.Contains(text, "#") {
		return
	}
	texts = strings.Fields(text)

	images, err := getImage(texts)
	if err != nil {
		logger.WithError(err).Error("获取图片失败")
		if errors.Is(err, errorNoCommand) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.NoCommand)
		} else if errors.Is(err, lolicon.ErrorKeywordNotFound) || errors.Is(err, pixiv.ErrorSearchFailed) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.KeywordNotFound)
		} else if errors.Is(err, lolicon.ErrorQuotaLimit) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.QuotaLimit)
		} else if errors.Is(err, errorTag) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.TagError)
		} else if errors.Is(err, errorNoTag) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.NoTagError)
		} else {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.GetImageFailed)
		}
		if len(images) == 0 {
			return
		}
	}
	sendPrivateImage(qqClient, msg.Sender.Uin, images)
}

func sendPrivateImage(qqClient *client.QQClient, qq int64, images [][]byte) {
	logger := logger.WithField("from", "sendPrivateImage")
	reply := message.NewSendingMessage()
	reply.Append(message.NewText(instance.config.Reply.Normal))
	num := 0
	for _, img := range images {
		if len(img) != 0 {
			r := bytes.NewReader(img)
			element, err := qqClient.UploadPrivateImage(qq, r)
			if err != nil {
				logger.WithError(err).Error("上传私聊图片失败")
				continue
			}
			reply.Append(element)
			num++
		}
	}
	if num != 0 {
		logger.Infof("给QQ %d 发送 %d 张图片", qq, num)
		if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
			logger.Errorf("给QQ %d 发送图片失败", qq)
			qqbot_utils.SendPrivateText(qqClient, qq, instance.config.Reply.SendImageFailed)
		}
	} else {
		qqbot_utils.SendPrivateText(qqClient, qq, instance.config.Reply.UploadImageFailed)
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
		text := strings.Join(texts, " ")
		if !strings.Contains(text, "#") {
			return
		}
		texts = strings.Fields(text)
		images, err := getImage(texts)
		if err != nil {
			logger.WithError(err).Error("获取图片失败")
			if errors.Is(err, errorNoCommand) {
				qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.NoCommand)
			} else if errors.Is(err, lolicon.ErrorKeywordNotFound) || errors.Is(err, pixiv.ErrorSearchFailed) {
				qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.KeywordNotFound)
			} else if errors.Is(err, lolicon.ErrorQuotaLimit) {
				qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.QuotaLimit)
			} else if errors.Is(err, errorTag) {
				qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.TagError)
			} else if errors.Is(err, errorNoTag) {
				qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.NoTagError)
			} else {
				qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.GetImageFailed)
			}
			if len(images) == 0 {
				return
			}
		}
		sendGroupImage(qqClient, msg, images)
	}
}

func sendGroupImage(qqClient *client.QQClient, msg *message.GroupMessage, images [][]byte) {
	logger := logger.WithField("from", "sendGroupImage")
	reply := message.NewSendingMessage()
	reply.Append(message.NewReply(msg))
	reply.Append(message.NewText(instance.config.Reply.Normal))
	num := 0
	for _, img := range images {
		if len(img) != 0 {
			r := bytes.NewReader(img)
			element, err := qqClient.UploadGroupImage(msg.GroupCode, r)
			if err != nil {
				logger.WithError(err).Error("上传群聊图片失败")
				continue
			}
			reply.Append(element)
			num++
		}
	}
	if num != 0 {
		logger.Infof("给QQ群 %d 里的QQ %d 发送 %d 张图片", msg.GroupCode, msg.Sender.Uin, num)
		if result := qqClient.SendGroupMessage(msg.GroupCode, reply); result == nil || result.Id <= 0 {
			logger.Errorf("给QQ群 %d 发送图片失败", msg.GroupCode)
			qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.SendImageFailed)
		}
	} else {
		qqbot_utils.SendGroupText(qqClient, msg, instance.config.Reply.UploadImageFailed)
	}
}

func getImage(texts []string) ([][]byte, error) {
	logger := logger.WithField("from", "getImage")

	var hasCommand bool
	keywords := make([]string, 0, len(texts))
	cmd := make(map[string]struct{})
	for _, t := range texts {
		var isCommand bool
		if strings.Contains(t, "#") {
			for k, v := range instance.commands {
				if strings.Contains(t, k) {
					hasCommand = true
					isCommand = true
					for _, c := range v {
						cmd[c] = struct{}{}
					}
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
	var e error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for s := range cmd {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()

			switch s {
			case lolicon.ID:
				query := instance.config.Lolicon
				if query.Keyword != "" {
					keywords = append(keywords, query.Keyword)
				}
				query.Keyword = strings.Join(keywords, " ")
				img, err := query.GetImage()
				if err != nil {
					logger.WithError(err).Error("获取lolicon图片失败")
					mu.Lock()
					e = err
					mu.Unlock()
					break
				}
				mu.Lock()
				images = append(images, img...)
				mu.Unlock()
			case islandwind233.AnimeID:
				anime := &islandwind233.Anime{}
				img, err := anime.GetImage()
				if err != nil {
					logger.WithError(err).Error("获取anime图片失败")
					mu.Lock()
					e = err
					mu.Unlock()
					break
				}
				mu.Lock()
				images = append(images, img...)
				mu.Unlock()
			case islandwind233.CosplayID:
				cosplay := &islandwind233.Cosplay{}
				img, err := cosplay.GetImage()
				if err != nil {
					logger.WithError(err).Error("获取cos图片失败")
					mu.Lock()
					e = err
					mu.Unlock()
					break
				}
				mu.Lock()
				images = append(images, img...)
				mu.Unlock()
			case paulzzh.ID:
				query := instance.config.Paulzzh
				if query.Tag != "" {
					keywords = append(keywords, query.Tag)
				}
				query.Tag = strings.Join(keywords, " ")
				if !setu_utils.IsLetter(query.Tag) {
					logger.WithError(errorTag).Errorf("错误Tag：%s", query.Tag)
					mu.Lock()
					e = errorTag
					mu.Unlock()
					break
				}
				img, err := query.GetImage()
				if err != nil {
					logger.WithError(err).Error("获取东方图片失败")
					mu.Lock()
					e = err
					mu.Unlock()
					break
				}
				mu.Lock()
				images = append(images, img...)
				mu.Unlock()
			case pixiv.ID:
				if len(keywords) == 0 {
					logger.WithError(errorNoTag).Error("没有关键字")
					mu.Lock()
					e = errorNoTag
					mu.Unlock()
					break
				}
				pixiv := *instance.pixiv
				pixiv.Tags = strings.Join(keywords, " ")
				img, err := pixiv.GetImage()
				if err != nil {
					logger.WithError(err).Error("获取pixiv图片失败")
					mu.Lock()
					e = err
					mu.Unlock()
					break
				}
				mu.Lock()
				images = append(images, img...)
				mu.Unlock()
			default:
				logger.Warnf("未知的图片接口：%s", s)
			}
		}(s)
	}
	wg.Wait()

	if len(images) == 0 {
		if e != nil {
			return nil, e
		}
		return nil, fmt.Errorf("获取图片失败")
	}

	return images, e
}

func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
