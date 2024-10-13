package main

import (
	"flag"
	"fmt"
	"github.com/kontsevoye/xkeen-control/confighandler"
	"github.com/kontsevoye/xkeen-control/xkeenipc"
	"gopkg.in/telebot.v3"
	telebotMiddleware "gopkg.in/telebot.v3/middleware"
	"log"
	"net/url"
	"strings"
	"time"
)

type appConfig struct {
	ConfigFilePath   string
	TelegramBotToken string
	TelegramAdminId  int64
}

type inlineAction string

const (
	add                  inlineAction = "add"
	remove               inlineAction = "remove"
	addReload            inlineAction = "addReload"
	removeReload         inlineAction = "removeReload"
	domainPrefix         inlineAction = "domainPrefix"
	exactPrefix          inlineAction = "exactPrefix"
	regexpPrefix         inlineAction = "regexpPrefix"
	withoutPrefix        inlineAction = "withoutPrefix"
	v2flyCommunityPrefix inlineAction = "v2flyCommunityPrefix"
)

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

	bot.Handle("/help", func(c telebot.Context) error {
		return c.Send(`
- Plain string: If this string matches any part of the target domain, the rule takes effect. For example, "sina.com" can match "sina.com", "sina.com.cn", and "www.sina.com", but not "sina.cn".
- Regular expression: Starts with "regexp:" followed by a regular expression. When this regular expression matches the target domain, the rule takes effect. For example, "regexp:\\.goo.*\\.com$" matches "www.google.com" or "fonts.googleapis.com", but not "google.com".
- Subdomain (recommended): Starts with "domain:" followed by a domain. When this domain is the target domain or a subdomain of the target domain, the rule takes effect. For example, "domain:xray.com" matches "www.xray.com" and "xray.com", but not "wxray.com".
- Exact match: Starts with "full:" followed by a domain. When this domain is an exact match for the target domain, the rule takes effect. For example, "full:xray.com" matches "xray.com" but not "www.xray.com".
- Load domains from a file: Formatted as "ext:file:tag", where the file is stored in the resource directory and has the same format as geosite.dat. The tag must exist in the file.
`)
	})

	dynamicHandler := func(inputText string) (string, *telebot.ReplyMarkup) {
		newMessageText := inputText
		inlineButtonRows := make([]telebot.Row, 0)
		reply := &telebot.ReplyMarkup{}

		if strings.HasPrefix(newMessageText, "http://") || strings.HasPrefix(newMessageText, "https://") {
			parsedUrl, err := url.Parse(newMessageText)
			if err != nil || parsedUrl.Host == "" {
				fmt.Printf("Хуйню суешь вместо хоста %s\n", newMessageText)
			} else {
				newMessageText = parsedUrl.Host
				hostSlice := strings.Split(newMessageText, ":")
				if len(hostSlice) == 2 {
					newMessageText = hostSlice[0]
				}
			}
		}

		prefixLessNewMessageText := newMessageText
		if strings.HasPrefix(newMessageText, "domain:") ||
			strings.HasPrefix(newMessageText, "full:") ||
			strings.HasPrefix(newMessageText, "regexp:") ||
			strings.HasPrefix(newMessageText, "ext:geosite_v2fly.dat:") {
			prefixLessNewMessageText = strings.Replace(prefixLessNewMessageText, "domain:", "", 1)
			prefixLessNewMessageText = strings.Replace(prefixLessNewMessageText, "full:", "", 1)
			prefixLessNewMessageText = strings.Replace(prefixLessNewMessageText, "regexp:", "", 1)
			prefixLessNewMessageText = strings.Replace(prefixLessNewMessageText, "ext:geosite_v2fly.dat:", "", 1)
			inlineButtonRows = append(
				inlineButtonRows,
				reply.Row(reply.Data(prefixLessNewMessageText, "", string(withoutPrefix), prefixLessNewMessageText)),
			)
		}
		if !strings.HasPrefix(newMessageText, "domain:") {
			inlineButtonRows = append(
				inlineButtonRows,
				reply.Row(reply.Data("domain:"+prefixLessNewMessageText, "", string(domainPrefix), "domain:"+prefixLessNewMessageText)),
			)
		}
		if !strings.HasPrefix(newMessageText, "full:") {
			inlineButtonRows = append(
				inlineButtonRows,
				reply.Row(reply.Data("full:"+prefixLessNewMessageText, "", string(exactPrefix), "full:"+prefixLessNewMessageText)),
			)
		}
		if !strings.HasPrefix(newMessageText, "regexp:") {
			inlineButtonRows = append(
				inlineButtonRows,
				reply.Row(reply.Data("regexp:"+prefixLessNewMessageText, "", string(regexpPrefix), "regexp:"+prefixLessNewMessageText)),
			)
		}
		if !strings.HasPrefix(newMessageText, "ext:geosite_v2fly.dat:") {
			inlineButtonRows = append(
				inlineButtonRows,
				reply.Row(reply.Data("ext:geosite_v2fly.dat:"+prefixLessNewMessageText, "", string(v2flyCommunityPrefix), "ext:geosite_v2fly.dat:"+prefixLessNewMessageText)),
			)
		}

		inlineButtonRows = append(
			inlineButtonRows,
			reply.Row(reply.Data("✅ Добавить", "", string(add), newMessageText)),
		)
		inlineButtonRows = append(
			inlineButtonRows,
			reply.Row(reply.Data("✅🔄 Добавить с перезапуском", "", string(addReload), newMessageText)),
		)
		inlineButtonRows = append(
			inlineButtonRows,
			reply.Row(reply.Data("⛔️ Удалить", "", string(remove), newMessageText)),
		)
		inlineButtonRows = append(
			inlineButtonRows,
			reply.Row(reply.Data("⛔️🔄 Удалить с перезапуском", "", string(removeReload), newMessageText)),
		)
		reply.Inline(inlineButtonRows...)

		return newMessageText, reply
	}

	bot.Handle(telebot.OnText, func(c telebot.Context) error {
		newMessageText, reply := dynamicHandler(c.Message().Text)

		return c.Send(newMessageText, reply)
	})

	bot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		dataSplit := strings.Split(c.Data(), "|")
		if len(dataSplit) != 2 {
			return c.Send("Какая-то залупа с колбеком")
		}
		var action inlineAction
		action = inlineAction(dataSplit[0])
		actionPayload := dataSplit[1]
		if action == add {
			err = confighandler.AddDomain(appConfig.ConfigFilePath, actionPayload)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка сохранения домена: %v", err))
			}
			_, err = c.Bot().Edit(c.Message(), fmt.Sprintf("✅ %s", actionPayload))
		} else if action == remove {
			err = confighandler.DeleteDomain(appConfig.ConfigFilePath, actionPayload)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка удаления домена: %v", err))
			}
			_, err = c.Bot().Edit(c.Message(), fmt.Sprintf("⛔️ %s", actionPayload))
		} else if action == addReload {
			err = confighandler.AddDomain(appConfig.ConfigFilePath, actionPayload)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка сохранения домена: %v", err))
			}
			err = xkeenipc.Restart()
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка перезапуска xkeen: %v", err))
			}
			_, err = c.Bot().Edit(c.Message(), fmt.Sprintf("✅🔄 %s", actionPayload))
		} else if action == removeReload {
			err = confighandler.DeleteDomain(appConfig.ConfigFilePath, actionPayload)
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка удаления домена: %v", err))
			}
			err = xkeenipc.Restart()
			if err != nil {
				return c.Send(fmt.Sprintf("Ошибка перезапуска xkeen: %v", err))
			}
			_, err = c.Bot().Edit(c.Message(), fmt.Sprintf("⛔🔄 %s", actionPayload))
		} else if action == domainPrefix ||
			action == exactPrefix ||
			action == regexpPrefix ||
			action == withoutPrefix ||
			action == v2flyCommunityPrefix {
			newMessageText, reply := dynamicHandler(actionPayload)
			_, err = c.Bot().Edit(c.Message(), newMessageText, reply)
		} else {
			return c.Send(fmt.Sprintf("Неизвестный экшон: %s (%s)", action, actionPayload))
		}

		return err
	})

	log.Println("Бот готов к запуску")
	bot.Start()
}
