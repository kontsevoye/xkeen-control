package main

import (
	"flag"
	"github.com/kontsevoye/xkeen-control/internal/confighandler"
	"github.com/kontsevoye/xkeen-control/internal/executor"
	"github.com/kontsevoye/xkeen-control/internal/telegrambotui"
	"github.com/kontsevoye/xkeen-control/internal/xkeenipc"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type appConfig struct {
	ConfigFilePath   string
	TelegramBotToken string
	TelegramAdminId  int64
	EnableBackups    bool
}

func newAppConfig(logger *zap.Logger) *appConfig {
	configFilePath := flag.String("config", "", "Путь к файлу конфигурации")
	telegramBotToken := flag.String("token", "", "Токен Telegram бота")
	telegramAdminId := flag.Int64("admin", 0, "Telegram ID админа")
	enableBackups := flag.Bool("enableBackups", true, "Включить бэкапы конфигов")

	flag.Parse()
	if *configFilePath == "" || *telegramBotToken == "" || *telegramAdminId == 0 {
		logger.Fatal(
			"missing required options config/token/admin",
			zap.String("configFilePath", *configFilePath),
			zap.String("telegramBotToken", *telegramBotToken),
			zap.Int64("telegramAdminId", *telegramAdminId),
		)
	}

	conf := &appConfig{
		ConfigFilePath:   *configFilePath,
		TelegramBotToken: *telegramBotToken,
		TelegramAdminId:  *telegramAdminId,
		EnableBackups:    *enableBackups,
	}

	return conf
}

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		err := logger.Sync()
		if err != nil && !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			panic(err)
		}
	}()
	conf := newAppConfig(logger)

	tui := telegrambotui.New(
		conf.TelegramBotToken,
		conf.TelegramAdminId,
		logger,
		confighandler.New(conf.ConfigFilePath, conf.EnableBackups),
		xkeenipc.New(executor.NewExecExecutor()),
	)
	go tui.Start()

	logger.Info("ready", zap.Int("pid", os.Getpid()))

	signalChan := make(chan os.Signal, 1)
	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
		syscall.SIGKILL,
	)
	s := <-signalChan
	tui.Stop()
	logger.Info("signal received, shutting down", zap.String("signal", s.String()))
}
