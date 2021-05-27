package command

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/utils"
	"github.com/Mrs4s/MiraiGo/client"
	"github.com/Mrs4s/MiraiGo/message"
	"github.com/orzogc/qqbot/command/moment"
	"github.com/orzogc/qqbot/command/moment/square"
	"github.com/orzogc/qqbot/command/search"
	"github.com/orzogc/qqbot/command/search/acfun"
	"github.com/orzogc/qqbot/command/search/duckduckgo"
	"github.com/orzogc/qqbot/command/search/ehentai"
	"github.com/orzogc/qqbot/command/search/google"
	"github.com/orzogc/qqbot/command/setu"
	"github.com/orzogc/qqbot/command/setu/islandwind233"
	"github.com/orzogc/qqbot/command/setu/lolicon"
	"github.com/orzogc/qqbot/command/setu/paulzzh"
	"github.com/orzogc/qqbot/command/setu/pixiv"
	"github.com/orzogc/qqbot/qqbot_utils"
	"github.com/spf13/viper"
)

const CommandID = "command" // ID

var (
	instance = &CommandBot{}                    // 机器人实例
	logger   = utils.GetModuleLogger(CommandID) // 日志记录
)

// 回复配置
type Reply struct {
	NoCommand string `json:"noCommand"` // 找不到命令的回复
}

// 配置
type Config struct {
	Setu   setu.Config   `json:"setu"`   // 图片机器人配置
	Search search.Config `json:"search"` // 搜索机器人配置
	Moment moment.Config `json:"moment"` // 动态机器人配置
	Reply  Reply         `json:"reply"`  // 回复配置
}

// 处理命令的机器人
type CommandBot struct {
	config   *Config                  // 配置
	commands map[string][]interface{} // 命令
	setubot  *setu.SetuBot            // 图片机器人
}

// 初始化
func init() {
	bot.RegisterModule(instance)
}

func (b *CommandBot) MiraiGoModule() bot.ModuleInfo {
	return bot.ModuleInfo{
		ID:       CommandID,
		Instance: instance,
	}
}

func (b *CommandBot) Init() {
	logger := logger.WithField("from", "Init")
	viper := viper.New()
	viper.SetConfigName(CommandID)
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

	instance.config.Setu.SetConfig()
	instance.config.Search.SetConfig()
	instance.config.Moment.SetConfig()
	if instance.config.Reply.NoCommand == "" {
		instance.config.Reply.NoCommand = "未知命令"
	}
	instance.setubot = setu.NewSetuBot(instance.config.Setu.Pixiv.PHPSESSID)

	setuCmd := map[string]setu.Setu{
		lolicon.LoliconID:       &instance.config.Setu.Lolicon,
		islandwind233.AnimeID:   &islandwind233.Anime{},
		islandwind233.CosplayID: &islandwind233.Cosplay{},
		paulzzh.PaulzzhID:       &instance.config.Setu.Paulzzh,
		pixiv.PixivID:           instance.setubot.Pixiv,
	}
	searchCmd := map[string]search.Search{
		google.GoogleID:         &google.Google{},
		duckduckgo.DuckDuckGoID: &duckduckgo.DuckDuckGo{},
		acfun.AcFunVideoID:      &acfun.AcFunVideo{},
		acfun.AcFunArticleID:    &acfun.AcFunArticle{},
		ehentai.EHentaiID:       &ehentai.EHentai{},
	}
	momentCmd := map[string]moment.Moment{
		square.AcFunSquareID: &square.AcFunSquare{},
	}
	instance.commands = make(map[string][]interface{})
	for k, v := range instance.config.Setu.Commands {
		command, ok := setuCmd[k]
		if !ok {
			logger.Warnf("未知的图片机器人命令ID：%s", k)
			continue
		}
		for _, s := range v {
			if c, ok := instance.commands[s]; ok {
				c = append(c, command)
				instance.commands[s] = c
			} else {
				instance.commands[s] = []interface{}{command}
			}
		}
	}
	for k, v := range instance.config.Search.Commands {
		command, ok := searchCmd[k]
		if !ok {
			logger.Warnf("未知的搜索机器人命令ID：%s", k)
			continue
		}
		for _, s := range v {
			if c, ok := instance.commands[s]; ok {
				c = append(c, command)
				instance.commands[s] = c
			} else {
				instance.commands[s] = []interface{}{command}
			}
		}
	}
	for k, v := range instance.config.Moment.Commands {
		command, ok := momentCmd[k]
		if !ok {
			logger.Warnf("未知的动态机器人命令ID：%s", k)
			continue
		}
		for _, s := range v {
			if c, ok := instance.commands[s]; ok {
				c = append(c, command)
				instance.commands[s] = c
			} else {
				instance.commands[s] = []interface{}{command}
			}
		}
	}
}

func (b *CommandBot) PostInit() {}

func (b *CommandBot) Serve(bot *bot.Bot) {
	registerBot(bot)
}

func (b *CommandBot) Start(bot *bot.Bot) {}

func (b *CommandBot) Stop(bot *bot.Bot, wg *sync.WaitGroup) {
	defer wg.Done()
}

// 处理私聊
func onPrivateMessage(qqClient *client.QQClient, msg *message.PrivateMessage) {
	logger := logger.WithField("from", "onPrivateMessage")

	text := qqbot_utils.GetPrivateText(msg)
	if !strings.Contains(text, "#") {
		return
	}
	cmd, keyword, err := getCmdAndKeyword(text)
	if err != nil {
		logger.WithError(err).WithField("privateMessage", msg.ToString()).Error("处理命令失败")
		if errors.Is(err, qqbot_utils.ErrorNoCommand) {
			qqbot_utils.SendPrivateText(qqClient, msg, instance.config.Reply.NoCommand)
		}
		return
	}

	setuCmd := make(map[setu.Setu]struct{})
	searchCmd := make(map[search.Search]struct{})
	momentCmd := make(map[moment.Moment]struct{})
	for c := range cmd {
		switch c := c.(type) {
		case setu.Setu:
			setuCmd[c] = struct{}{}
		case search.Search:
			searchCmd[c] = struct{}{}
		case moment.Moment:
			momentCmd[c] = struct{}{}
		default:
			logger.Warn("未知的机器人命令类型")
		}
	}

	if len(setuCmd) != 0 {
		instance.config.Setu.HandlePrivateMessage(qqClient, msg, setuCmd, keyword)
	}
	if len(searchCmd) != 0 {
		instance.config.Search.HandlePrivateMessage(qqClient, msg, searchCmd, keyword)
	}
	if len(momentCmd) != 0 {
		instance.config.Moment.HandlePrivateMessage(qqClient, msg, momentCmd, keyword)
	}
}

// 处理群聊
func onGroupMessage(qqClient *client.QQClient, msg *message.GroupMessage) {
	logger := logger.WithField("from", "onGroupMessage")

	if text, isAt := qqbot_utils.GetGroupAtText(qqClient.Uin, msg); isAt {
		if !strings.Contains(text, "#") {
			return
		}
		cmd, keyword, err := getCmdAndKeyword(text)
		if err != nil {
			logger.WithError(err).WithField("groupMessage", msg.ToString()).Error("处理命令失败")
			if errors.Is(err, qqbot_utils.ErrorNoCommand) {
				qqbot_utils.ReplyGroupText(qqClient, msg, instance.config.Reply.NoCommand)
			}
			return
		}

		setuCmd := make(map[setu.Setu]struct{})
		searchCmd := make(map[search.Search]struct{})
		momentCmd := make(map[moment.Moment]struct{})
		for c := range cmd {
			switch c := c.(type) {
			case setu.Setu:
				setuCmd[c] = struct{}{}
			case search.Search:
				searchCmd[c] = struct{}{}
			case moment.Moment:
				momentCmd[c] = struct{}{}
			default:
				logger.Warn("未知的机器人命令类型")
			}
		}

		if len(setuCmd) != 0 {
			instance.config.Setu.HandleGroupMessage(qqClient, msg, setuCmd, keyword)
		}
		if len(searchCmd) != 0 {
			instance.config.Search.HandleGroupMessage(qqClient, msg, searchCmd, keyword)
		}
		if len(momentCmd) != 0 {
			instance.config.Moment.HandleGroupMessage(qqClient, msg, momentCmd, keyword)
		}
	}
}

// 处理命令和关键词
func getCmdAndKeyword(text string) (cmd map[interface{}]struct{}, keyword string, err error) {
	texts := strings.Fields(text)
	var hasCommand bool
	keywords := make([]string, 0, len(texts))
	cmd = make(map[interface{}]struct{})
	for _, t := range texts {
		var isCommand bool
		if strings.Contains(t, "#") {
			for k, v := range instance.commands {
				if strings.Contains(strings.ToLower(t), k) {
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
		return nil, "", qqbot_utils.ErrorNoCommand
	}

	return cmd, strings.Join(keywords, " "), nil
}

// 注册mirai事件函数
func registerBot(b *bot.Bot) {
	b.OnPrivateMessage(onPrivateMessage)
	b.OnGroupMessage(onGroupMessage)
}
