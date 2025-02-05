package localization

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	messages       map[string]map[string]string
	defaultLang    = "en"                 // Язык по умолчанию
	supportedLangs = []string{"en", "ru"} // Поддерживаемые языки
	mu             sync.Mutex
)

func LoadMessages() {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.ReadFile("./configs/messages.json")
	if err != nil {
		log.Panic(err)
	}

	if err := json.Unmarshal(file, &messages); err != nil {
		log.Panic(err)
	}
}

// getMessage возвращает сообщение для указанного языка
func GetMessage(lang, key string, args ...interface{}) string {
	mu.Lock()
	defer mu.Unlock()

	// Если язык не поддерживается, используем язык по умолчанию
	if _, ok := messages[lang]; !ok {
		lang = defaultLang
	}

	msg, ok := messages[lang][key]
	if !ok {
		return "Message not found."
	}

	if len(args) > 0 {
		return fmt.Sprintf(msg, args...)
	}
	return msg
}

// getUserLang определяет язык пользователя
func GetUserLang(langCode string) string {
	for _, lang := range supportedLangs {
		if lang == langCode {
			return lang
		}
	}
	return defaultLang // Если язык не поддерживается
}
