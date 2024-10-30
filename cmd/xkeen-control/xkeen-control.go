package main

import (
	"github.com/kontsevoye/xkeen-control/internal/config"
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

func main() {
	logger, _ := zap.NewProduction()
	defer func() {
		err := logger.Sync()
		if err != nil && !strings.Contains(err.Error(), "inappropriate ioctl for device") {
			panic(err)
		}
	}()
	conf := config.New(logger)

	tui := telegrambotui.New(
		conf.TelegramBotToken,
		conf.TelegramAdminId,
		logger,
		confighandler.New(conf.ConfigFilePath, conf.EnableBackups),
		xkeenipc.New(executor.NewExecExecutor(), conf.XkeenBinaryName),
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
