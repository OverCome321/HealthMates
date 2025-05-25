# Используем официальный Go-образ как базовый
FROM golang:1.21-alpine

# Устанавливаем рабочую директорию
WORKDIR /app

# Устанавливаем необходимые пакеты для сборки
RUN apk add --no-cache gcc musl-dev

# Копируем go.mod и go.sum
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod download && \
    go mod verify

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Открываем порт для сервера
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
