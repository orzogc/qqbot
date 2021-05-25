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
	"github.com/spf13/viper"
)

// ID
const SetuID = "setu"

var (
	instance       = &SetuBot{}                    // 机器人实例
	logger         = utils.GetModuleLogger(SetuID) // 日志记录
	errorNoCommand = errors.New("没有发现有效的命令")
)

// 回复配置
type Reply struct {
	Normal            string `json:"normal"`            // 正常回复
	NoCommand         string `json:"noCommand"`         // 找不到命令的回复
	GetImageFailed    string `json:"getImageFailed"`    // 获取图片失败的回复
	UploadImageFailed string `json:"uploadImageFailed"` // 上传图片失败的回复
	SendImageFailed   string `json:"sendImageFailed"`   // 发送图片是版的回复
	KeywordNotFound   string `json:"keywordNotFound"`   // 搜索图片失败的回复
	QuotaLimit        string `json:"quotaLimit"`        // 达到接口额度的回复
	TagError          string `json:"tagError"`          // 搜索关键字错误的回复
	NoTagError        string `json:"noTagError"`        // 没有搜索关键字的回复
}

// pixiv配置
type PixivConfig struct {
	PHPSESSID    string             `json:"phpsessid"`    // pixiv网页Cookie里的PHPSESSID，为空的话没有r18图片
	SearchOption pixiv.SearchOption `json:"searchOption"` // 搜索选项
}

// 配置
type Config struct {
	Lolicon  lolicon.Lolicon     `json:"lolicon"`  // lolicon的配置，keyword是无效的
	Paulzzh  paulzzh.Paulzzh     `json:"paulzzh"`  // paulzzh的配置，tag是无效的
	Pixiv    PixivConfig         `json:"pixiv"`    // pixiv的配置
	Commands map[string][]string `json:"commands"` // 命令关键字
	Reply    Reply               `json:"reply"`    // 回复配置
}

// 图片机器人
type SetuBot struct {
	config   *Config
	pixiv    *pixiv.Pixiv
	commands map[string][]Setu
}

// 初始化
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
	} else {
		err = viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件setu.json的内容无效，使用默认设置")
			instance.config = new(Config)
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

	cmd := map[string]Setu{
		lolicon.ID:              &instance.config.Lolicon,
		islandwind233.AnimeID:   &islandwind233.Anime{},
		islandwind233.CosplayID: &islandwind233.Cosplay{},
		paulzzh.ID:              &instance.config.Paulzzh,
		pixiv.ID:                instance.pixiv,
	}
	instance.commands = make(map[string][]Setu)
	for k, v := range instance.config.Commands {
		setu, ok := cmd[k]
		if !ok {
			logger.Warnf("未知的命令ID：%s", k)
			continue
		}
		for _, s := range v {
			if c, ok := instance.commands[s]; ok {
				c = append(c, setu)
				instance.commands[s] = c
			} else {
				instance.commands[s] = []Setu{setu}
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

// 处理私聊
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
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("获取图片失败")
		if errors.Is(err, errorNoCommand) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.NoCommand)
		} else if errors.Is(err, lolicon.ErrorKeywordNotFound) || errors.Is(err, pixiv.ErrorSearchFailed) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.KeywordNotFound)
		} else if errors.Is(err, lolicon.ErrorQuotaLimit) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.QuotaLimit)
		} else if errors.Is(err, paulzzh.ErrorTag) {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.TagError)
		} else if errors.Is(err, pixiv.ErrorNoTag) {
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

// 发送私聊图片
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
				logger.WithError(err).WithField("receiverQQ", qq).Error("上传私聊图片失败")
				continue
			}
			reply.Append(element)
			num++
		}
	}
	if num != 0 {
		logger.WithField("receiverQQ", qq).Infof("发送 %d 张私聊图片", num)
		if result := qqClient.SendPrivateMessage(qq, reply); result == nil || result.Id <= 0 {
			logger.WithField("receiverQQ", qq).Error("发送私聊图片失败")
			qqbot_utils.SendPrivateText(qqClient, qq, instance.config.Reply.SendImageFailed)
		}
	} else {
		qqbot_utils.SendPrivateText(qqClient, qq, instance.config.Reply.UploadImageFailed)
	}
}

// 处理群聊
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
			logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("获取图片失败")
			if errors.Is(err, errorNoCommand) {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.NoCommand)
			} else if errors.Is(err, lolicon.ErrorKeywordNotFound) || errors.Is(err, pixiv.ErrorSearchFailed) {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.KeywordNotFound)
			} else if errors.Is(err, lolicon.ErrorQuotaLimit) {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.QuotaLimit)
			} else if errors.Is(err, paulzzh.ErrorTag) {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.TagError)
			} else if errors.Is(err, pixiv.ErrorNoTag) {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.NoTagError)
			} else {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.GetImageFailed)
			}
			if len(images) == 0 {
				return
			}
		}
		sendGroupImage(qqClient, msg, images)
	}
}

// 发送群聊图片
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
				logger.WithError(err).
					WithField("qqGroup", msg.GroupCode).
					WithField("receiverQQ", msg.Sender.Uin).
					Error("上传群聊图片失败")
				continue
			}
			reply.Append(element)
			num++
		}
	}
	if num != 0 {
		logger.WithField("qqGroup", msg.GroupCode).
			WithField("receiverQQ", msg.Sender.Uin).
			Infof("发送 %d 张群聊图片", num)
		if result := qqClient.SendGroupMessage(msg.GroupCode, reply); result == nil || result.Id <= 0 {
			logger.
				WithField("qqGroup", msg.GroupCode).
				WithField("receiverQQ", msg.Sender.Uin).
				Error("发送群聊图片失败")
			qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.SendImageFailed)
		}
	} else {
		qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.UploadImageFailed)
	}
}

// 获取图片
func getImage(texts []string) ([][]byte, error) {
	logger := logger.WithField("from", "getImage")

	var hasCommand bool
	keywords := make([]string, 0, len(texts))
	cmd := make(map[Setu]struct{})
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
	keyword := strings.Join(keywords, " ")

	var images [][]byte
	var e error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for s := range cmd {
		wg.Add(1)
		go func(s Setu) {
			defer wg.Done()
			img, err := s.GetImage(keyword)
			if err != nil {
				logger.WithError(err).Error("获取图片失败")
				mu.Lock()
				defer mu.Unlock()
				e = err
				return
			}
			mu.Lock()
			defer mu.Unlock()
			images = append(images, img...)
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

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
