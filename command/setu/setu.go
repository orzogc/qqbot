package setu

import (
	"bytes"
	"errors"
	"fmt"
	"sync"

	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/setu/islandwind233"
	"github.com/orzogc/qqbot/command/setu/lolicon"
	"github.com/orzogc/qqbot/command/setu/paulzzh"
	"github.com/orzogc/qqbot/command/setu/pixiv"
	"github.com/orzogc/qqbot/command/setu/setu_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const SetuID = "setu" // ID

var logger = utils.GetModuleLogger(SetuID) // 日志记录

var (
	_ Setu = (*lolicon.Lolicon)(nil)
	_ Setu = (*islandwind233.Anime)(nil)
	_ Setu = (*islandwind233.Cosplay)(nil)
	_ Setu = (*paulzzh.Paulzzh)(nil)
	_ Setu = (*pixiv.Pixiv)(nil)
)

// 图片接口
type Setu interface {
	// 获取图片，keyword为搜索关键字，可以不支持搜索，可返回多个图片
	GetImage(keyword string) (*setu_utils.Image, error)
}

// 回复配置
type Reply struct {
	Normal            string `json:"normal"`            // 正常回复
	GetImageFailed    string `json:"getImageFailed"`    // 获取图片失败的回复
	UploadImageFailed string `json:"uploadImageFailed"` // 上传图片失败的回复
	SendImageFailed   string `json:"sendImageFailed"`   // 发送图片失败的回复
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
	config *Config      // 配置
	pixiv  *pixiv.Pixiv // pixiv
}

// 新建图片机器人
func NewSetuBot(config *Config) *SetuBot {
	return &SetuBot{
		config: config,
		pixiv:  pixiv.New(config.Pixiv.PHPSESSID),
	}
}

// 部分配置没有设置的话采用默认配置，设置cmd，返回设置好的cmd，实现Command接口
func (b *SetuBot) SetConfig(cmd map[string][]interface{}) map[string][]interface{} {
	if len(b.config.Commands) == 0 {
		b.config.Commands = map[string][]string{
			lolicon.LoliconID:       {"色图", "涩图", "瑟图", "setu"},
			islandwind233.AnimeID:   {"二次元", "二刺猿", "erciyuan"},
			islandwind233.CosplayID: {"cos", "余弦", "三次元"},
			paulzzh.PaulzzhID:       {"东方", "车万", "東方", "越共", "dongfang", "touhou"},
			pixiv.PixivID:           {"pixiv", "p站"},
		}
	}
	if b.config.Reply.Normal == "" {
		b.config.Reply.Normal = "这是您点的图片"
	}
	if b.config.Reply.GetImageFailed == "" {
		b.config.Reply.GetImageFailed = "获取图片失败"
	}
	if b.config.Reply.UploadImageFailed == "" {
		b.config.Reply.UploadImageFailed = "上传图片失败"
	}
	if b.config.Reply.SendImageFailed == "" {
		b.config.Reply.SendImageFailed = "发送图片失败"
	}
	if b.config.Reply.KeywordNotFound == "" {
		b.config.Reply.KeywordNotFound = "找不到符合关键字的图片"
	}
	if b.config.Reply.QuotaLimit == "" {
		b.config.Reply.QuotaLimit = "已经达到接口的调用额度上限"
	}
	if b.config.Reply.TagError == "" {
		b.config.Reply.TagError = "东方图片搜索关键字必须是英文字母"
	}
	if b.config.Reply.NoTagError == "" {
		b.config.Reply.NoTagError = "pixiv图片搜索需要关键字"
	}

	setuCmd := map[string]Setu{
		lolicon.LoliconID:       &b.config.Lolicon,
		islandwind233.AnimeID:   &islandwind233.Anime{},
		islandwind233.CosplayID: &islandwind233.Cosplay{},
		paulzzh.PaulzzhID:       &b.config.Paulzzh,
		pixiv.PixivID:           b.pixiv,
	}
	for k, v := range b.config.Commands {
		command, ok := setuCmd[k]
		if !ok {
			logger.Warnf("未知的图片机器人命令ID：%s", k)
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
func (b *SetuBot) HandlePrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, cmd map[interface{}]struct{}, keyword string) {
	logger := logger.WithField("from", "HandlePrivateMessage")

	setuCmd := make(map[Setu]struct{})
	for c := range cmd {
		if c, ok := c.(Setu); ok {
			setuCmd[c] = struct{}{}
		}
	}
	if len(setuCmd) == 0 {
		return
	}

	images, err := getImage(setuCmd, keyword)
	if err != nil {
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("获取图片失败")
		if errors.Is(err, lolicon.ErrorKeywordNotFound) || errors.Is(err, pixiv.ErrorSearchFailed) {
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.KeywordNotFound)
		} else if errors.Is(err, lolicon.ErrorQuotaLimit) {
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.QuotaLimit)
		} else if errors.Is(err, paulzzh.ErrorTag) {
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.TagError)
		} else if errors.Is(err, pixiv.ErrorNoTag) {
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.NoTagError)
		} else {
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.GetImageFailed)
		}
		if len(images) == 0 {
			return
		}
	}
	if len(images) != 0 {
		b.sendPrivateImage(qqClient, msg, images)
	}
}

// 处理群聊消息，实现Command接口
func (b *SetuBot) HandleGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, cmd map[interface{}]struct{}, keyword string) {
	logger := logger.WithField("from", "HandleGroupMessage")

	setuCmd := make(map[Setu]struct{})
	for c := range cmd {
		if c, ok := c.(Setu); ok {
			setuCmd[c] = struct{}{}
		}
	}
	if len(setuCmd) == 0 {
		return
	}

	images, err := getImage(setuCmd, keyword)
	if err != nil {
		logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("获取图片失败")
		if errors.Is(err, lolicon.ErrorKeywordNotFound) || errors.Is(err, pixiv.ErrorSearchFailed) {
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.KeywordNotFound)
		} else if errors.Is(err, lolicon.ErrorQuotaLimit) {
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.QuotaLimit)
		} else if errors.Is(err, paulzzh.ErrorTag) {
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.TagError)
		} else if errors.Is(err, pixiv.ErrorNoTag) {
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.NoTagError)
		} else {
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.GetImageFailed)
		}
		if len(images) == 0 {
			return
		}
	}
	if len(images) != 0 {
		b.sendGroupImage(qqClient, msg, images)
	}
}

// 发送私聊图片
func (b *SetuBot) sendPrivateImage(qqClient *client.QQClient, msg *message.PrivateMessage, images []*setu_utils.Image) {
	logger := logger.WithField("from", "sendPrivateImage")

	reply := message.NewSendingMessage()
	reply.Append(message.NewText(b.config.Reply.Normal + "\n"))
	num := 0
	for _, image := range images {
		if image.Text != "" && len(image.Images) != 0 {
			reply.Append(message.NewText(image.Text + "\n"))
		}
		for _, img := range image.Images {
			if len(img) != 0 {
				element, err := qqClient.UploadPrivateImage(msg.Sender.Uin, bytes.NewReader(img))
				if err != nil {
					logger.WithError(err).WithField("receiverQQ", msg.Sender.Uin).Error("上传私聊图片失败")
					continue
				}
				reply.Append(element)
				num++
			}
		}
	}
	if num != 0 {
		logger.WithField("receiverQQ", msg.Sender.Uin).Infof("发送 %d 张私聊图片", num)
		if result := qqClient.SendPrivateMessage(msg.Sender.Uin, reply); result == nil || result.Id <= 0 {
			logger.WithField("receiverQQ", msg.Sender.Uin).Error("发送私聊图片失败")
			qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.SendImageFailed)
		}
	} else {
		qqbot_utils.SendPrivateText(qqClient, msg, b.config.Reply.UploadImageFailed)
	}
}

// 发送群聊图片
func (b *SetuBot) sendGroupImage(qqClient *client.QQClient, msg *message.GroupMessage, images []*setu_utils.Image) {
	logger := logger.WithField("from", "sendGroupImage")

	reply := message.NewSendingMessage()
	reply.Append(message.NewReply(msg))
	reply.Append(message.NewText(b.config.Reply.Normal + "\n"))
	num := 0
	for _, image := range images {
		if image.Text != "" && len(image.Images) != 0 {
			reply.Append(message.NewText(image.Text + "\n"))
		}
		for _, img := range image.Images {
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
				num++
			}
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
			qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.SendImageFailed)
		}
	} else {
		qqbot_utils.ReplyGroupText(qqClient, msg, b.config.Reply.UploadImageFailed)
	}
}

// 获取图片
func getImage(cmd map[Setu]struct{}, keyword string) ([]*setu_utils.Image, error) {
	logger := logger.WithField("from", "getImage")

	images := make([]*setu_utils.Image, 0, len(cmd))
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
			images = append(images, img)
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
