package search

import (
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
	"github.com/orzogc/qqbot/search/acfun"
	"github.com/orzogc/qqbot/search/duckduckgo"
	"github.com/orzogc/qqbot/search/google"
	"github.com/orzogc/qqbot/search/search_utils"
	"github.com/spf13/viper"
)

const SearchID = "search" // ID

var (
	instance = &SearchBot{}                    // 机器人实例
	logger   = utils.GetModuleLogger(SearchID) // 日志记录
)

// 回复配置
type Reply struct {
	SearchFailed     string `json:"searchFailed"`
	SendResultFailed string `json:"sendResultFailed"`
}

// 配置
type Config struct {
	Commands map[string][]string `json:"commands"` // 命令关键字
	Reply    Reply               `json:"reply"`    // 回复配置
}

// 负责网页搜索的bot
type SearchBot struct {
	config        *Config             // 配置
	commands      map[string][]Search // 命令
	otherCommands map[string]struct{} // 其他机器人的命令
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
	logger := logger.WithField("from", "Init")
	viper := viper.New()
	viper.SetConfigName(SearchID)
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
		logger.WithError(err).Warn("读取设置文件search.json失败，使用默认设置")
		instance.config = new(Config)
	} else {
		err = viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件search.json的内容无效，使用默认设置")
			instance.config = new(Config)
		}
	}

	if len(instance.config.Commands) == 0 {
		instance.config.Commands = map[string][]string{
			google.GoogleID:         {"google", "谷歌"},
			duckduckgo.DuckDuckGoID: {"duck"},
			acfun.AcFunVideoID:      {"ac", "a站", "缺b乐", "缺逼乐", "爱稀饭"},
			acfun.AcFunArticleID:    {"文章"},
		}
	}
	if instance.config.Reply.SearchFailed == "" {
		instance.config.Reply.SearchFailed = "搜索失败"
	}
	if instance.config.Reply.SendResultFailed == "" {
		instance.config.Reply.SendResultFailed = "发送搜索结果失败"
	}

	cmd := map[string]Search{
		google.GoogleID:         &google.Google{},
		duckduckgo.DuckDuckGoID: &duckduckgo.DuckDuckGo{},
		acfun.AcFunVideoID:      &acfun.AcFunVideo{},
		acfun.AcFunArticleID:    &acfun.AcFunArticle{},
	}
	instance.commands = make(map[string][]Search)
	instance.otherCommands = make(map[string]struct{})
	for k, v := range instance.config.Commands {
		search, ok := cmd[k]
		if !ok {
			logger.Warnf("未知的命令ID：%s", k)
			continue
		}
		for _, s := range v {
			if c, ok := instance.commands[s]; ok {
				c = append(c, search)
				instance.commands[s] = c
			} else {
				instance.commands[s] = []Search{search}
			}
			qqbot_utils.AllCommands[s] = struct{}{}
		}
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
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("搜索失败")
		if errors.Is(err, qqbot_utils.ErrorNoCommand) {
			//qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, "未知命令")
		} else {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.SearchFailed)
		}
		if result == "" {
			return
		}
	}
	if result != "" {
		if ok := qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, result); !ok {
			qqbot_utils.SendPrivateText(qqClient, msg.Sender.Uin, instance.config.Reply.SendResultFailed)
		}
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
			logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("搜索失败")
			if errors.Is(err, qqbot_utils.ErrorNoCommand) {
				//qqbot_utils.ReplyGroupText(qqClient, msg, "未知命令")
			} else {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.SearchFailed)
			}
			if result == "" {
				return
			}
		}
		if result != "" {
			if ok := qqbot_utils.ReplyGroupText(qqClient, msg, result); !ok {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.SendResultFailed)
			}
		}
	}
}

// 搜索
func search(text string) (string, error) {
	logger := logger.WithField("from", "search")

	texts := strings.Fields(text)
	var hasCommand bool
	var hasOtherCommand bool
	keywords := make([]string, 0, len(texts))
	cmd := make(map[Search]struct{})
	for _, t := range texts {
		s := strings.ToLower(t)
		var isCommand bool
		if strings.Contains(t, "#") {
			for k, v := range instance.commands {
				if strings.Contains(s, k) {
					hasCommand = true
					isCommand = true
					for _, c := range v {
						cmd[c] = struct{}{}
					}
				}
			}
			for k := range instance.otherCommands {
				if strings.Contains(s, k) {
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

	return search_utils.ConvertToText(result), nil
}

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
