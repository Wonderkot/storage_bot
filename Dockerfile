# Используем официальный образ Go
FROM golang:1.23.5-alpine as golang

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем исходный код в контейнер
COPY . .

# Собираем приложение
RUN go mod download
RUN go mod verify
RUN CGO_ENABLED=0 GOOS=linux go build -o /bot ./cmd/bot

# Указываем точку входа
ENTRYPOINT ["/bot"]