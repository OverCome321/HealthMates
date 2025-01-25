# Используем официальный Go-образ как базовый
FROM golang:1.23-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для загрузки зависимостей
COPY go.mod go.sum ./

# Загружаем зависимости
RUN go mod tidy

# Копируем весь исходный код приложения
COPY . .

# Компилируем приложение
RUN go build -o main .

# Используем минимальный образ для запуска приложения
FROM alpine:latest

# Устанавливаем необходимые пакеты для работы с приложением
RUN apk --no-cache add ca-certificates

# Копируем собранный бинарник из этапа сборки
COPY --from=builder /app/main /app/main

# Устанавливаем рабочую директорию
WORKDIR /app

# Открываем порт для сервера
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
