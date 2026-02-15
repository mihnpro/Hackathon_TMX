# services/task3/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Установка зависимостей
RUN apk add --no-cache git gcc musl-dev

# Копирование модулей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Сборка приложения
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/task3-app ./cmd/web

# Финальный образ
FROM alpine:latest

WORKDIR /app

# Установка curl
RUN apk add --no-cache curl ca-certificates tzdata

# Копирование собранного приложения
COPY --from=builder /app/bin/task3-app /app/task3-app

# Копирование фронтенда и данных
COPY --from=builder /app/frontend /app/frontend
COPY --from=builder /app/data /app/data
COPY --from=builder /app/internal /app/internal

# Создание директорий
RUN mkdir -p /app/uploads /app/maps

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

CMD ["/app/task3-app"]