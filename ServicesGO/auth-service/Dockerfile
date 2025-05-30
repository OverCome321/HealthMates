FROM golang:1.24-alpine AS builder
WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -o auth-service cmd/app/main.go

# --- 2) Runtime stage ---
FROM alpine:latest
WORKDIR /app

# Копируем бинарь и .env
COPY --from=builder /app/auth-service .
COPY .env ./

# Копируем swagger-спецификации
COPY internal/transport/rest/docs /app/docs

# Открываем порт
EXPOSE 8080

ENTRYPOINT ["./auth-service"]
