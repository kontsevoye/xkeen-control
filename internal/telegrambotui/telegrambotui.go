package telegrambotui

import (
	"fmt"
	"github.com/kontsevoye/xkeen-control/internal/confighandler"
	"github.com/kontsevoye/xkeen-control/internal/xkeenipc"
	"go.uber.org/zap"
	"gopkg.in/telebot.v3"
	telebotMiddleware "gopkg.in/telebot.v3/middleware"
	"net/url"
	"strings"
	"time"
)

func New(
	telegramBotToken string,
	telegramAdminId int64,
	logger *zap.Logger,
	configHandler *confighandler.ConfigHandler,
	xkeenIpc *xkeenipc.XkeenIpc,
) *TelegramBotUi {
	tui := &TelegramBotUi{
		telegramBotToken: telegramBotToken,
		telegramAdminId:  telegramAdminId,
		logger:           logger,
		configHandler:    configHandler,
		xkeenIpc:         xkeenIpc,
	}
	tui.initialize()

	return tui
}

type TelegramBotUi struct {
	telegramBot      *telebot.Bot
	telegramBotToken string
	telegramAdminId  int64
	logger           *zap.Logger
	configHandler    *confighandler.ConfigHandler
	xkeenIpc         *xkeenipc.XkeenIpc
}

type inlineAction string

// TODO: Callback data has a 64 byte limit, solve this in a more correct way
const (
	add                  inlineAction = "add"
	remove               inlineAction = "rm"
	addReload            inlineAction = "addRl"
	removeReload         inlineAction = "rmRl"
	domainPrefix         inlineAction = "dp"
	exactPrefix          inlineAction = "ep"
	regexpPrefix         inlineAction = "rp"
	withoutPrefix        inlineAction = "wp"
	v2flyCommunityPrefix inlineAction = "cp"
	cutSubdomain         inlineAction = "cs"
)

func (tui *TelegramBotUi) loggerMiddleware() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			fields := []zap.Field{
				zap.Int("update_id", c.Update().ID),
				zap.String("recipient", c.Recipient().Recipient()),
				zap.String("text", c.Text()),
				zap.String("data", c.Data()),
			}
			if c.Message() != nil {
				fields = append(fields, zap.Int("message_id", c.Message().ID))
				fields = append(fields, zap.Int64("sender_id", c.Message().Sender.ID))
			}
			if c.Callback() != nil {
				fields = append(fields, zap.String("callback_id", c.Callback().ID))
			}
			tui.logger.Info("tg update", fields...)
			defer func() {
				err := tui.logger.Sync()
				if err != nil && !strings.Contains(err.Error(), "inappropriate ioctl for device") {
					fmt.Println(err)
				}
			}()
			err := next(c)
			if err != nil {
				fields = append(fields, zap.Error(err))
				tui.logger.Error("handle error", fields...)
			}

			return err
		}
	}
}

func (tui *TelegramBotUi) escapeTgMarkdownSpecialCharacters(input string) string {
	specialCharacters := []string{
		"_",
		"*",
		"[",
		"]",
		"(",
		")",
		"~",
		"`",
		">",
		"#",
		"+",
		"-",
		"=",
		"|",
		"{",
		"}",
		".",
		"!",
	}
	for _, specialCharacter := range specialCharacters {
		input = strings.ReplaceAll(input, specialCharacter, fmt.Sprintf("\\%s", specialCharacter))
	}

	return input
}

func (tui *TelegramBotUi) initialize() {
	bot, err := telebot.NewBot(telebot.Settings{
		Token:  tui.telegramBotToken,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		tui.logger.Fatal("cant create telegram bot", zap.Error(err))
	}
	tui.telegramBot = bot

	tui.telegramBot.Use(tui.loggerMiddleware())
	tui.telegramBot.Use(telebotMiddleware.Whitelist(tui.telegramAdminId))

	tui.telegramBot.Handle("/list", func(c telebot.Context) error {
		domains, err := tui.configHandler.GetDomains()
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка чтения доменов: %v", err))
		}
		if len(domains) == 0 {
			return c.Send("Список доменов пуст.")
		}
		for i, domain := range domains {
			domains[i] = fmt.Sprintf("`%s`", tui.escapeTgMarkdownSpecialCharacters(domain))
		}

		return c.Send(
			fmt.Sprintf("*Текущий список доменов:*\n\\- %s", strings.Join(domains, "\n\\- ")),
			telebot.ModeMarkdownV2,
		)
	})

	tui.telegramBot.Handle("/add", func(c telebot.Context) error {
		domain := c.Message().Payload
		if domain == "" {
			return c.Send("Пожалуйста, укажите домен для добавления.")
		}

		err = tui.configHandler.AddDomain(domain)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка сохранения домена: %v", err))
		}

		return c.Send("Домен успешно добавлен.")
	})

	tui.telegramBot.Handle("/delete", func(c telebot.Context) error {
		domain := c.Message().Payload
		if domain == "" {
			return c.Send("Пожалуйста, укажите домен для удаления.")
		}

		err = tui.configHandler.DeleteDomain(domain)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка удаления домена: %v", err))
		}

		return c.Send("Домен успешно удален.")
	})

	tui.telegramBot.Handle("/restart", func(c telebot.Context) error {
		err = c.Send("начинаю перезапуск")
		if err != nil {
			return err
		}
		err = tui.xkeenIpc.Restart()
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка перезапуска xkeen: %v", err))
		}

		return c.Send("xkeen успешно перезапущен.")
	})

	tui.telegramBot.Handle("/backups", func(c telebot.Context) error {
		backupFiles, err := tui.configHandler.ListBackupFiles()
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка получения списка бэкапов: %v", err))
		}
		if len(backupFiles) == 0 {
			return c.Send("Список бэкапов пуст.")
		}

		return c.Send(fmt.Sprintf("Текущий список бэкапов:\n- %s", strings.Join(backupFiles, "\n- ")))
	})

	tui.telegramBot.Handle("/restore", func(c telebot.Context) error {
		backupFileName := c.Message().Payload
		if backupFileName == "" {
			return c.Send("Пожалуйста, укажите файл для восстановления.")
		}

		err = tui.configHandler.RestoreBackup(backupFileName)
		if err != nil {
			return c.Send(fmt.Sprintf("Ошибка восстановления из бэкапа: %v", err))
		}

		return c.Send("Бэкап успешно восстановлен.")
	})

	tui.telegramBot.Handle("/help", func(c telebot.Context) error {
		return c.Send(strings.ReplaceAll(strings.ReplaceAll(`
>\- Plain string: If this string matches any part of the target domain, the rule takes effect\. For example, "==sina\.com==" can match "==sina\.com==", "==sina\.com\.cn==", and "==www\.sina\.com==", but not "==sina\.cn=="\.
>\- Regular expression: Starts with "==regexp:==" followed by a regular expression\. When this regular expression matches the target domain, the rule takes effect\. For example, "==regexp:\\\\\.goo\.\*\\\\\.com$==" matches "==www\.google\.com==" or "==fonts\.googleapis\.com==", but not "==google\.com=="\.
>\- Subdomain \(recommended\): Starts with "==domain:==" followed by a domain\. When this domain is the target domain or a subdomain of the target domain, the rule takes effect\. For example, "==domain:xray\.com==" matches "==www\.xray\.com==" and "==xray\.com==", but not "==wxray\.com=="\.
>\- Exact match: Starts with "==full:==" followed by a domain\. When this domain is an exact match for the target domain, the rule takes effect\. For example, "==full:xray\.com==" matches "==xray\.com==" but not "==www\.xray\.com=="\.
>\- Load domains from a file: Formatted as "==ext:file:tag==", where the file is stored in the resource directory and has the same format as ==geosite\.dat==\. The tag must exist in the file\.
`, "\n", "\n>\n"), "==", "`"), telebot.ModeMarkdownV2)
	})

	dynamicHandler := func(inputText string) (string, *telebot.ReplyMarkup) {
		newMessageText := inputText
		inlineButtonRows := make([]telebot.Row, 0)
		reply := &telebot.ReplyMarkup{}

		if strings.HasPrefix(newMessageText, "http://") || strings.HasPrefix(newMessageText, "https://") {
			parsedUrl, err := url.Parse(newMessageText)
			if err != nil || parsedUrl.Host == "" {
				tui.logger.Warn("zalupa instead of host", zap.String("newMessageText", newMessageText))
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
		if !strings.HasPrefix(newMessageText, "domain:") &&
			!strings.HasPrefix(newMessageText, "full:") &&
			!strings.HasPrefix(newMessageText, "regexp:") &&
			!strings.HasPrefix(newMessageText, "ext:geosite_v2fly.dat:") {
			subdomains := strings.Split(prefixLessNewMessageText, ".")
			if len(subdomains) > 2 {
				parent := strings.Join(subdomains[1:], ".")
				tui.logger.Info(parent)
				inlineButtonRows = append(
					inlineButtonRows,
					reply.Row(reply.Data(parent, "", string(cutSubdomain), parent)),
				)
			}
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
			reply.Row(
				reply.Data("✅ Добавить", "", string(add), newMessageText),
				reply.Data("✅🔄 с перезапуском", "", string(addReload), newMessageText),
			),
		)
		inlineButtonRows = append(
			inlineButtonRows,
			reply.Row(
				reply.Data("⛔️ Удалить", "", string(remove), newMessageText),
				reply.Data("⛔️🔄 с перезапуском", "", string(removeReload), newMessageText),
			),
		)
		reply.Inline(inlineButtonRows...)

		return fmt.Sprintf("`%s`", tui.escapeTgMarkdownSpecialCharacters(newMessageText)), reply
	}

	tui.telegramBot.Handle(telebot.OnText, func(c telebot.Context) error {
		newMessageText, reply := dynamicHandler(c.Message().Text)

		return c.Send(
			newMessageText,
			reply,
			telebot.ModeMarkdownV2,
		)
	})

	tui.telegramBot.Handle(telebot.OnCallback, func(c telebot.Context) error {
		dataSplit := strings.Split(c.Data(), "|")
		if len(dataSplit) != 2 {
			return c.Send("Какая-то залупа с колбеком")
		}
		var action inlineAction
		action = inlineAction(dataSplit[0])
		actionPayload := dataSplit[1]
		if action == add {
			err = tui.configHandler.AddDomain(actionPayload)
			if err != nil {
				_, err = c.Bot().Edit(
					c.Message(),
					fmt.Sprintf(
						"⚠️ Ошибка сохранения \"`%s`\": %s",
						tui.escapeTgMarkdownSpecialCharacters(actionPayload),
						tui.escapeTgMarkdownSpecialCharacters(err.Error()),
					),
					telebot.ModeMarkdownV2,
				)
				return err
			}
			_, err = c.Bot().Edit(
				c.Message(),
				fmt.Sprintf("✅ `%s`", tui.escapeTgMarkdownSpecialCharacters(actionPayload)),
				telebot.ModeMarkdownV2,
			)
		} else if action == remove {
			err = tui.configHandler.DeleteDomain(actionPayload)
			if err != nil {
				_, err = c.Bot().Edit(
					c.Message(),
					fmt.Sprintf(
						"⚠️ Ошибка удаления \"`%s`\": %s",
						tui.escapeTgMarkdownSpecialCharacters(actionPayload),
						tui.escapeTgMarkdownSpecialCharacters(err.Error()),
					),
					telebot.ModeMarkdownV2,
				)
				return err
			}
			_, err = c.Bot().Edit(
				c.Message(),
				fmt.Sprintf("⛔️ `%s`", tui.escapeTgMarkdownSpecialCharacters(actionPayload)),
				telebot.ModeMarkdownV2,
			)
		} else if action == addReload {
			err = tui.configHandler.AddDomain(actionPayload)
			if err != nil {
				_, err = c.Bot().Edit(
					c.Message(),
					fmt.Sprintf(
						"⚠️ Ошибка сохранения \"`%s`\": %s",
						tui.escapeTgMarkdownSpecialCharacters(actionPayload),
						tui.escapeTgMarkdownSpecialCharacters(err.Error()),
					),
					telebot.ModeMarkdownV2,
				)
				return err
			}
			_, err = c.Bot().Edit(
				c.Message(),
				fmt.Sprintf("✅ `%s`\n🔄 Перезапускаю xkeen...", tui.escapeTgMarkdownSpecialCharacters(actionPayload)),
				telebot.ModeMarkdownV2,
			)
			err = tui.xkeenIpc.Restart()
			if err != nil {
				_, err = c.Bot().Edit(
					c.Message(),
					fmt.Sprintf(
						"⚠️ Ошибка перезапуска xkeen \"`%s`\": %s\n✅ Но в список добавлен😬",
						tui.escapeTgMarkdownSpecialCharacters(actionPayload),
						tui.escapeTgMarkdownSpecialCharacters(err.Error()),
					),
					telebot.ModeMarkdownV2,
				)
				return err
			}
			_, err = c.Bot().Edit(
				c.Message(),
				fmt.Sprintf("✅🔄 `%s`", tui.escapeTgMarkdownSpecialCharacters(actionPayload)),
				telebot.ModeMarkdownV2,
			)
		} else if action == removeReload {
			err = tui.configHandler.DeleteDomain(actionPayload)
			if err != nil {
				_, err = c.Bot().Edit(
					c.Message(),
					fmt.Sprintf(
						"⚠️ Ошибка удаления \"`%s`\": %s",
						tui.escapeTgMarkdownSpecialCharacters(actionPayload),
						tui.escapeTgMarkdownSpecialCharacters(err.Error()),
					),
					telebot.ModeMarkdownV2,
				)
				return err
			}
			_, err = c.Bot().Edit(
				c.Message(),
				fmt.Sprintf("⛔ `%s`\n🔄 Перезапускаю xkeen...", tui.escapeTgMarkdownSpecialCharacters(actionPayload)),
				telebot.ModeMarkdownV2,
			)
			err = tui.xkeenIpc.Restart()
			if err != nil {
				_, err = c.Bot().Edit(
					c.Message(),
					fmt.Sprintf(
						"⚠️ Ошибка перезапуска xkeen \"`%s`\": %s\n⛔ Но из списка удален😬",
						tui.escapeTgMarkdownSpecialCharacters(actionPayload),
						tui.escapeTgMarkdownSpecialCharacters(err.Error()),
					),
					telebot.ModeMarkdownV2,
				)
				return err
			}
			_, err = c.Bot().Edit(
				c.Message(),
				fmt.Sprintf("⛔🔄 `%s`", tui.escapeTgMarkdownSpecialCharacters(actionPayload)),
				telebot.ModeMarkdownV2,
			)
		} else if action == domainPrefix ||
			action == exactPrefix ||
			action == regexpPrefix ||
			action == withoutPrefix ||
			action == v2flyCommunityPrefix ||
			action == cutSubdomain {
			newMessageText, reply := dynamicHandler(actionPayload)
			_, err = c.Bot().Edit(c.Message(), newMessageText, reply, telebot.ModeMarkdownV2)
		} else {
			return c.Send(fmt.Sprintf("Неизвестный экшон: %s (%s)", action, actionPayload))
		}

		return err
	})
}

func (tui *TelegramBotUi) Start() {
	tui.telegramBot.Start()
}

func (tui *TelegramBotUi) Stop() {
	tui.telegramBot.Stop()
}
