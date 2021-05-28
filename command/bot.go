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
	"github.com/orzogc/qqbot/command/search"
	"github.com/orzogc/qqbot/command/setu"
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
	config    *Config                  // 配置
	commands  map[string][]interface{} // 命令
	setuBot   *setu.SetuBot            // 图片机器人
	searchBot *search.SearchBot        // 搜索机器人
	momentBot *moment.MomentBot        // 动态机器人
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
		logger.WithError(err).Warn("读取设置文件command.json失败，使用默认设置")
		instance.config = new(Config)
	} else {
		err = viper.Unmarshal(&instance.config)
		if err != nil {
			logger.WithError(err).Warn("设置文件command.json的内容无效，使用默认设置")
			instance.config = new(Config)
		}
	}

	if instance.config.Reply.NoCommand == "" {
		instance.config.Reply.NoCommand = "未知命令"
	}
	instance.commands = make(map[string][]interface{})
	instance.setuBot = setu.NewSetuBot(&instance.config.Setu)
	instance.searchBot = search.NewSearchBot(&instance.config.Search)
	instance.momentBot = moment.NewMomentBot(&instance.config.Moment)
	instance.commands = instance.setuBot.SetConfig(instance.commands)
	instance.commands = instance.searchBot.SetConfig(instance.commands)
	instance.commands = instance.momentBot.SetConfig(instance.commands)
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

	if len(cmd) != 0 {
		instance.setuBot.HandlePrivateMessage(qqClient, msg, cmd, keyword)
		instance.searchBot.HandlePrivateMessage(qqClient, msg, cmd, keyword)
		instance.momentBot.HandlePrivateMessage(qqClient, msg, cmd, keyword)
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

		if len(cmd) != 0 {
			instance.setuBot.HandleGroupMessage(qqClient, msg, cmd, keyword)
			instance.searchBot.HandleGroupMessage(qqClient, msg, cmd, keyword)
			instance.momentBot.HandleGroupMessage(qqClient, msg, cmd, keyword)
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
