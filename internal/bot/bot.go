package bot

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"storage_bot/internal/localization"
	"storage_bot/internal/storage"
	"storage_bot/internal/whitelist"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// startBot запускает бота
func StartBot() {
	// Загружаем текстовые сообщения
	localization.LoadMessages()

	// Загружаем белый список
	whitelist.LoadWhitelist()

	// Загружаем данные из файла при старте
	storage.InitStorage()
	// Получаем токен бота из переменной окружения
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Panic("TELEGRAM_BOT_TOKEN не установлен")
	}

	// Определяем, включен ли белый список
	enableWhitelist, _ := strconv.ParseBool(os.Getenv("ENABLE_WHITELIST"))

	// Определяем, включен ли режим отладки
	debugMode, _ := strconv.ParseBool(os.Getenv("DEBUG_MODE"))

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	// Включаем или выключаем режим отладки
	bot.Debug = debugMode

	// Устанавливаем вебхук
	webhookURL := os.Getenv("WEBHOOK_URL")
	if webhookURL == "" {
		log.Panic("WEBHOOK_URL не установлен")
	}

	hook, _ := tgbotapi.NewWebhook(webhookURL)

	_, err = bot.Request(hook)
	if err != nil {
		log.Panic(err)
	}

	adminid, err := strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	if err != nil {
		log.Panic("ADMIN_ID не установлен")
	}

	updates := bot.ListenForWebhook("/")

	go func() {
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	for update := range updates {
		if update.Message == nil { // Игнорируем любые не-сообщения
			continue
		}

		// Проверяем доступ пользователя, если белый список включен
		if enableWhitelist && !whitelist.IsUserAllowed(update.Message.From.ID) {
			continue // Пропускаем сообщения от пользователей не из белого списка
		}

		// Определяем язык пользователя
		lang := localization.GetUserLang(update.Message.From.LanguageCode)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "start":
				msg.Text = localization.GetMessage(lang, "start")
			case "list":
				list := storage.ListEntries()
				if list == "" {
					msg.Text = localization.GetMessage(lang, "list_empty")
				} else {
					msg.Text = list
				}
			case "help":
				msg.Text = localization.GetMessage(lang, "help")
			case "remove":
				key := strings.TrimSpace(update.Message.CommandArguments())
				if key == "" {
					msg.Text = localization.GetMessage(lang, "invalid_format")
				} else {
					removedKey, success := storage.RemoveEntry(key)
					if success {
						msg.Text = localization.GetMessage(lang, "remove_success", removedKey)
					} else {
						msg.Text = localization.GetMessage(lang, "remove_fail", removedKey)
					}
				}
			case "wipe":
				if update.Message.From.ID == adminid {
					storage.Wipe()
					msg.Text = "clear"

				} else {
					msg.Text = "not allowed"
				}
			default:
				msg.Text = localization.GetMessage(lang, "unknown_command")
			}
		} else {
			text := update.Message.Text

			// Валидируем и парсим сообщение
			key, value, err := validateAndParseMessage(text)
			if err != nil {
				msg.Text = localization.GetMessage(lang, "invalid_format")
			} else {
				storage.AddEntry(key, value)
				msg.Text = localization.GetMessage(lang, "saved", key, value)
			}
		}

		if _, err := bot.Send(msg); err != nil {
			log.Panic(err)
		}
	}
}

// validateAndParseMessage проверяет сообщение и возвращает ключ и значение
func validateAndParseMessage(text string) (string, string, error) {
	// Проверка на пустую строку
	if text == "" {
		return "", "", fmt.Errorf("сообщение пустое")
	}

	// Проверка на наличие разделителя
	parts := strings.SplitN(text, ":", 2)
	if len(parts) < 2 {
		return "", "", fmt.Errorf("неверный формат. Используйте key:value")
	}

	// Убираем пробелы по краям
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])

	// Проверка, что ключ и значение не пустые
	if key == "" || value == "" {
		return "", "", fmt.Errorf("ключ и значение не могут быть пустыми")
	}

	return key, value, nil
}
