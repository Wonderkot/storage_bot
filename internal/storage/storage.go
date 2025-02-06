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

// InitStorage инициализирует путь к файлу и загружает данные
func InitStorage() {
	mu.Lock()
	defer mu.Unlock()

	// Определяем путь к файлу
	storageFile = os.Getenv("STORAGE_FILE_PATH")
	if storageFile == "" {
		storageFile = "/app/internal/storage/data.json" // Дефолтный путь в контейнере
	}
	fmt.Println("Используем файл хранилища:", storageFile)

	// Проверяем, существует ли файл
	if _, err := os.Stat(storageFile); os.IsNotExist(err) {
		fmt.Println("Файл не найден, создаем новый:", storageFile)
		if err := saveEmptyStorage(); err != nil {
			log.Printf("Ошибка создания файла %s: %v\n", storageFile, err)
			return
		}
	}

	// Читаем файл
	file, err := os.ReadFile(storageFile)
	if err != nil {
		log.Printf("Ошибка чтения %s: %v\n", storageFile, err)
		return
	}

	// Декодируем JSON
	if err := json.Unmarshal(file, &storage); err != nil {
		log.Printf("Ошибка парсинга JSON (%s): %v\n", storageFile, err)
		return
	}

	fmt.Printf("Загружено %d записей из %s\n", len(storage), storageFile)
}

// saveEmptyStorage создаёт пустой JSON-файл, если его нет
func saveEmptyStorage() error {
	data := make(map[string]string)
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("ошибка сериализации JSON: %v", err)
	}

	if err := os.WriteFile(storageFile, jsonData, 0644); err != nil {
		return fmt.Errorf("ошибка записи в %s: %v", storageFile, err)
	}

	fmt.Println("Файл", storageFile, "успешно создан с пустым JSON {}")
	return nil
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
