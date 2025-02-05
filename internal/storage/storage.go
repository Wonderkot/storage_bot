package storage

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

var (
	storage     = make(map[string]string)
	storageFile = "data.json"
	mu          sync.Mutex
)

func LoadStorage() {
	mu.Lock()
	defer mu.Unlock()

	file, err := os.ReadFile(storageFile)
	if err != nil {
		if os.IsNotExist(err) {
			return // Файл не существует, начинаем с пустого хранилища
		}
		log.Panic(err)
	}

	if err := json.Unmarshal(file, &storage); err != nil {
		log.Panic(err)
	}
}

// listEntries возвращает список всех записей
func ListEntries() string {
	mu.Lock()
	defer mu.Unlock()

	if len(storage) == 0 {
		return ""
	}

	var result strings.Builder
	for key, value := range storage {
		result.WriteString(fmt.Sprintf("%s: %s\n", key, value))
	}
	return result.String()
}

// addEntry добавляет запись в хранилище
func AddEntry(key, value string) {
	storage[key] = value
	saveStorage()
}

// removeEntry удаляет запись по ключу
func RemoveEntry(key string) (string, bool) {
	if _, exists := storage[key]; exists {
		delete(storage, key)
		saveStorage()
		return key, true
	}
	return key, false
}

func Wipe() {
	for k := range storage {
		delete(storage, k)
	}
	saveStorage()
}

func saveStorage() {
	mu.Lock()
	defer mu.Unlock()

	file, err := json.MarshalIndent(storage, "", "  ")
	if err != nil {
		log.Panic(err)
	}

	if err := os.WriteFile(storageFile, file, 0644); err != nil {
		log.Panic(err)
	}
}
