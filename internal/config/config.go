package config

import (
	"flag"
	"go.uber.org/zap"
)

type AppConfig struct {
	ConfigFilePath   string
	TelegramBotToken string
	TelegramAdminId  int64
	EnableBackups    bool
	XkeenBinaryName  string
}

func New(logger *zap.Logger) *AppConfig {
	configFilePath := flag.String("config", "", "Путь к файлу конфигурации")
	telegramBotToken := flag.String("token", "", "Токен Telegram бота")
	telegramAdminId := flag.Int64("admin", 0, "Telegram ID админа")
	enableBackups := flag.Bool("enableBackups", true, "Включить бэкапы конфигов, опционально")
	xkeenBinaryName := flag.String("xkeenBinaryName", "", "Имя пакета xkeen, опционально")

	flag.Parse()
	if *configFilePath == "" || *telegramBotToken == "" || *telegramAdminId == 0 {
		logger.Fatal(
			"missing required options config/token/admin",
			zap.String("configFilePath", *configFilePath),
			zap.String("telegramBotToken", *telegramBotToken),
			zap.Int64("telegramAdminId", *telegramAdminId),
		)
	}

	conf := &AppConfig{
		ConfigFilePath:   *configFilePath,
		TelegramBotToken: *telegramBotToken,
		TelegramAdminId:  *telegramAdminId,
		EnableBackups:    *enableBackups,
		XkeenBinaryName:  *xkeenBinaryName,
	}

	return conf
}
