package whitelist

import (
	"encoding/json"
	"log"
	"os"
	"sync"
)

var (
	whitelist     []int64 // Список ID пользователей
	whitelistFile = "./configs/whitelist.json"
	mu            sync.Mutex
)

func LoadWhitelist() {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.ReadFile(whitelistFile)
	if err != nil {
		if os.IsNotExist(err) {
			return // Файл не существует, начинаем с пустого списка
		}
		log.Panic(err)
	}

	if err := json.Unmarshal(file, &whitelist); err != nil {
		log.Panic(err)
	}
}

// isUserAllowed проверяет, есть ли пользователь в белом списке
func IsUserAllowed(userID int64) bool {
	mu.Lock()
	defer mu.Unlock()

	for _, id := range whitelist {
		if id == userID {
			return true
		}
	}
	return false
}
