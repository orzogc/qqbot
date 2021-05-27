package search

import (
	"fmt"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/search/acfun"
	"github.com/orzogc/qqbot/command/search/duckduckgo"
	"github.com/orzogc/qqbot/command/search/ehentai"
	"github.com/orzogc/qqbot/command/search/google"
	"github.com/orzogc/qqbot/command/search/search_utils"
	"github.com/orzogc/qqbot/qqbot_utils"
)

const SearchID = "search" // ID

var logger = utils.GetModuleLogger(SearchID) // 日志记录

var _ Search = (*google.Google)(nil)
var _ Search = (*duckduckgo.DuckDuckGo)(nil)
var _ Search = (*acfun.AcFunVideo)(nil)
var _ Search = (*acfun.AcFunArticle)(nil)
var _ Search = (*ehentai.EHentai)(nil)

// 搜索接口
type Search interface {
	// 搜索text
	Search(text string) ([]search_utils.SearchResult, error)
}

// 回复配置
type Reply struct {
	SearchFailed     string `json:"searchFailed"`     // 搜索失败的回复
	SendResultFailed string `json:"sendResultFailed"` // 发送搜索结果失败的回复
}

// 配置
type Config struct {
	Commands map[string][]string `json:"commands"` // 命令关键字
	Reply    Reply               `json:"reply"`    // 回复配置
}

// 部分配置没有设置的话采用默认配置
func (c *Config) SetConfig() {
	if len(c.Commands) == 0 {
		c.Commands = map[string][]string{
			google.GoogleID:         {"google", "谷歌"},
			duckduckgo.DuckDuckGoID: {"duck"},
			acfun.AcFunVideoID:      {"ac", "a站", "缺b乐", "缺逼乐", "爱稀饭"},
			acfun.AcFunArticleID:    {"文章"},
			ehentai.EHentaiID:       {"eh"},
		}
	}
	if c.Reply.SearchFailed == "" {
		c.Reply.SearchFailed = "搜索失败"
	}
	if c.Reply.SendResultFailed == "" {
		c.Reply.SendResultFailed = "发送搜索结果失败"
	}
}

// 处理私聊消息
func (c *Config) HandlePrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage, cmd map[Search]struct{}, keyword string) {
	logger := logger.WithField("from", "HandlePrivateMessage")

	result, err := search(cmd, keyword)
	if err != nil {
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("搜索失败")
		qqbot_utils.SendPrivateText(qqClient, msg, c.Reply.SearchFailed)
		if result == "" {
			return
		}
	}
	if result != "" {
		if ok := qqbot_utils.SendPrivateText(qqClient, msg, result); !ok {
			qqbot_utils.SendPrivateText(qqClient, msg, c.Reply.SendResultFailed)
		}
	}
}

// 处理群聊消息
func (c *Config) HandleGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage, cmd map[Search]struct{}, keyword string) {
	logger := logger.WithField("from", "HandleGroupMessage")

	result, err := search(cmd, keyword)
	if err != nil {
		logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("搜索失败")
		qqbot_utils.ReplyGroupText(qqClient, msg, c.Reply.SearchFailed)
		if result == "" {
			return
		}
	}
	if result != "" {
		if ok := qqbot_utils.ReplyGroupText(qqClient, msg, result); !ok {
			qqbot_utils.ReplyGroupText(qqClient, msg, c.Reply.SendResultFailed)
		}
	}
}

// 搜索
func search(cmd map[Search]struct{}, keyword string) (string, error) {
	logger := logger.WithField("from", "search")

	if strings.TrimSpace(keyword) == "" {
		return "", fmt.Errorf("搜索关键字为空")
	}

	var result []search_utils.SearchResult
	var e error
	var mu sync.Mutex
	var wg sync.WaitGroup
	for s := range cmd {
		wg.Add(1)
		go func(s Search) {
			defer wg.Done()
			r, err := s.Search(keyword)
			if err != nil {
				logger.WithError(err).Error("搜索失败")
				mu.Lock()
				defer mu.Unlock()
				e = err
				return
			}
			mu.Lock()
			defer mu.Unlock()
			result = append(result, r...)
		}(s)
	}
	wg.Wait()

	if len(result) == 0 {
		if e != nil {
			return "", e
		}
		return "", fmt.Errorf("搜索失败")
	}

	return search_utils.ToString(result), nil
}
