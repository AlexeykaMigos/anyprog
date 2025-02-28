FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем и загружаем зависимости
COPY go.mod go.sum ./
# Копируем исходный код и .env файл
RUN go mod download

COPY . .
COPY .env .env

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o myapp .

FROM alpine:latest

WORKDIR /root/

# Копируем бинарный файл и .env файл
COPY --from=builder /app/myapp .
COPY --from=builder /app/.env .env

# Открываем порт для приложения
EXPOSE 8080

# Запускаем приложение
CMD ["./myapp"]