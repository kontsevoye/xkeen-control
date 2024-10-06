package main

import (
	"flag"
	"fmt"
	"github.com/kontsevoye/xkeen-control/confighandler"
	"github.com/kontsevoye/xkeen-control/xkeenipc"
	"gopkg.in/telebot.v3"
	telebotMiddleware "gopkg.in/telebot.v3/middleware"
	"log"
	"strings"
	"time"
)

type appConfig struct {
	ConfigFilePath   string
	TelegramBotToken string
	TelegramAdminId  int64
}

func newAppConfig() *appConfig {
	configFilePath := flag.String("config", "", "Путь к файлу конфигурации")
	telegramBotToken := flag.String("token", "", "Токен Telegram бота")
	telegramAdminId := flag.Int64("admin", 0, "Telegram ID админа")

	flag.Parse()
	if *configFilePath == "" || *telegramBotToken == "" || *telegramAdminId == 0 {
		log.Fatal("Пожалуйста, укажите путь к файлу конфигурации с помощью -config, токен бота с помощью -token, ID админа с помощью -admin")
	}

	appConfig := &appConfig{
		ConfigFilePath:   *configFilePath,
		TelegramBotToken: *telegramBotToken,
		TelegramAdminId:  *telegramAdminId,
	}

	return appConfig
}

func main() {
	appConfig := newAppConfig()

	bot, err := telebot.NewBot(telebot.Settings{
		Token:  appConfig.TelegramBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatal("Ошибка создания бота:", err)
	}

	bot.Use(telebotMiddleware.Logger())
	bot.Use(telebotMiddleware.Whitelist(appConfig.TelegramAdminId))

	bot.Handle("/list", func(c telebot.Context) error {
		domains, err := confighandler.GetDomains(appConfig.ConfigFilePath)
		if err != nil {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Ошибка чтения доменов: %v", err))
			return nil
		}
		if len(domains) == 0 {
			c.Bot().Send(c.Message().Sender, "Список доменов пуст.")
		} else {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Текущий список доменов:\n- %s", strings.Join(domains, "\n- ")))
		}

		return nil
	})

	bot.Handle("/add", func(c telebot.Context) error {
		domain := c.Message().Payload
		if domain == "" {
			c.Bot().Send(c.Message().Sender, "Пожалуйста, укажите домен для добавления.")
			return nil
		}

		err = confighandler.AddDomain(appConfig.ConfigFilePath, domain)
		if err != nil {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Ошибка сохранения домена: %v", err))
			return nil
		}
		c.Bot().Send(c.Message().Sender, "Домен успешно добавлен.")

		return nil
	})

	bot.Handle("/delete", func(c telebot.Context) error {
		domain := c.Message().Payload
		if domain == "" {
			c.Bot().Send(c.Message().Sender, "Пожалуйста, укажите домен для удаления.")
			return nil
		}

		err = confighandler.DeleteDomain(appConfig.ConfigFilePath, domain)
		if err != nil {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Ошибка удаления домена: %v", err))
			return nil
		}
		c.Bot().Send(c.Message().Sender, "Домен успешно удален.")

		return nil
	})

	bot.Handle("/restart", func(c telebot.Context) error {
		err = xkeenipc.Restart()
		if err != nil {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Ошибка перезапуска xkeen: %v", err))
			return nil
		}
		c.Bot().Send(c.Message().Sender, "xkeen успешно перезапущен.")

		return nil
	})

	bot.Handle("/backups", func(c telebot.Context) error {
		backupFiles, err := confighandler.ListBackupFiles(appConfig.ConfigFilePath)
		if err != nil {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Ошибка получения списка бэкапов: %v", err))
			return nil
		}
		if len(backupFiles) == 0 {
			c.Bot().Send(c.Message().Sender, "Список бэкапов пуст.")
		} else {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Текущий список бэкапов:\n- %s", strings.Join(backupFiles, "\n- ")))
		}

		return nil
	})

	bot.Handle("/restore", func(c telebot.Context) error {
		backupFileName := c.Message().Payload
		if backupFileName == "" {
			c.Bot().Send(c.Message().Sender, "Пожалуйста, укажите файл для восстановления.")
			return nil
		}

		err = confighandler.RestoreBackup(appConfig.ConfigFilePath, backupFileName)
		if err != nil {
			c.Bot().Send(c.Message().Sender, fmt.Sprintf("Ошибка восстановления из бэкапа: %v", err))
			return nil
		}
		c.Bot().Send(c.Message().Sender, "Бэкап успешно восстановлен.")

		return nil
	})

	log.Println("Бот готов к запуску")
	bot.Start()
}
