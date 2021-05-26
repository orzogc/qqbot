package search

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/orzogc/qqbot/search/google"
	"github.com/orzogc/qqbot/search/search_utils"
)

const SearchID = "search" // ID

var (
	instance = &SearchBot{}                    // 机器人实例
	logger   = utils.GetModuleLogger(SearchID) // 日志记录
)

// 负责网页搜索的bot
type SearchBot struct {
	commands      map[string]Search
	otherCommands map[string]struct{}
}

// 初始化
func init() {
	bot.RegisterModule(instance)
}

func (b *SearchBot) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       SearchID,
		Instance: instance,
	}
}

func (b *SearchBot) Init() {
	instance.commands = map[string]Search{
		google.GoogleID: &google.Google{},
	}
	instance.otherCommands = make(map[string]struct{})
	for c := range instance.commands {
		qqbot_utils.AllCommands[c] = struct{}{}
	}
}

func (b *SearchBot) PostInit() {
	for c := range qqbot_utils.AllCommands {
		if _, ok := instance.commands[c]; !ok {
			instance.otherCommands[c] = struct{}{}
		}
	}
}

func (b *SearchBot) Serve(bot *bot.Bot) {
	registerBot(bot)
}

func (b *SearchBot) Start(bot *bot.Bot) {}

func (b *SearchBot) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

// 处理私聊
func onPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage) {
	logger := logger.WithField("from", "onPrivateMessage")

	text := qqbot_utils.GetPrivateText(msg)
	if !strings.Contains(text, "#") {
		return
	}

	result, err := search(text)
	if err != nil {
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("搜索网页失败")
		if errors.Is(err, qqbot_utils.ErrorNoCommand) {
			//qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, "未知命令")
		} else {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, "搜索网页失败")
		}
		if result == "" {
			return
		}
	}
	if result != "" {
		qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, result)
	}
}

// 处理群聊
func onGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage) {
	logger := logger.WithField("from", "onGroupMessage")

	if text, isAt := qqbot_utils.GetGroupAtText(qqClient.Uin, msg); isAt {
		if !strings.Contains(text, "#") {
			return
		}

		result, err := search(text)
		if err != nil {
			logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("搜索网页失败")
			if errors.Is(err, qqbot_utils.ErrorNoCommand) {
				//qqbot_utils.ReplyGroupText(qqClient, msg, "未知命令")
			} else {
				qqbot_utils.ReplyGroupText(qqClient, msg, "搜索网页失败")
			}
			if result == "" {
				return
			}
		}
		if result != "" {
			qqbot_utils.ReplyGroupText(qqClient, msg, result)
		}
	}
}

// 搜索网页
func search(text string) (string, error) {
	logger := logger.WithField("from", "search")

	texts := strings.Fields(text)
	var hasCommand bool
	var hasOtherCommand bool
	keywords := make([]string, 0, len(texts))
	cmd := make(map[Search]struct{})
	for _, t := range texts {
		var isCommand bool
		if strings.Contains(t, "#") {
			for k, v := range instance.commands {
				if strings.Contains(t, k) {
					hasCommand = true
					isCommand = true
					cmd[v] = struct{}{}
				}
			}
			for k := range instance.otherCommands {
				if strings.Contains(t, k) {
					hasOtherCommand = true
					isCommand = true
				}
			}
		}
		if !isCommand {
			keywords = append(keywords, t)
		}
	}
	if !hasCommand && hasOtherCommand {
		return "", nil
	}
	if !hasCommand && !hasOtherCommand {
		return "", qqbot_utils.ErrorNoCommand
	}
	keyword := strings.Join(keywords, " ")

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
				logger.WithError(err).Error("搜索网页失败")
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
		return "", fmt.Errorf("搜索网页失败")
	}

	return search_utils.ConvertToText(result), nil
}

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
