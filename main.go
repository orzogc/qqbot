package main

import (
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/Logiase/MiraiGo-Template/bot"
	"github.com/Logiase/MiraiGo-Template/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	_ "github.com/orzogc/qqbot/achan"
	_ "github.com/orzogc/qqbot/command"
	_ "github.com/orzogc/qqbot/logmessage"
	"github.com/orzogc/qqbot/qqbot_utils"
	_ "github.com/orzogc/qqbot/reconnect"
)

// 读取配置
func loadConfig() {
	logger := logrus.WithField("from", "loadConfig")

	config.GlobalConfig = &config.Config{
		Viper: viper.New(),
	}
	config.GlobalConfig.SetConfigName("qqbot")
	config.GlobalConfig.SetConfigType("json")
	path, err := os.Executable()
	if err != nil {
		logger.WithError(err).Panic("获取执行文件所在位置失败")
	}

	dir := filepath.Dir(path)
	config.GlobalConfig.AddConfigPath(dir)
	config.GlobalConfig.AddConfigPath(".")

	err = config.GlobalConfig.ReadInConfig()
	if err != nil {
		logger.WithError(err).Panic("读取设置文件qqbot.json失败")
	}

	timeout := config.GlobalConfig.GetUint("timeout")
	if timeout != 0 {
		qqbot_utils.Timeout = timeout
	}
	qqbot_utils.SetTimeout(qqbot_utils.Timeout)
}

func main() {
	genDevice := flag.Bool("g", false, "生成随机设备文件device.json")
	flag.Parse()
	if *genDevice {
		bot.GenRandomDevice()
		return
	}

	loadConfig()

	bot.Init()
	bot.StartService()
	bot.UseProtocol(bot.AndroidPhone)
	bot.Login()
	bot.RefreshList()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-ch
	signal.Stop(ch)
	signal.Reset(os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	bot.Stop()
}
