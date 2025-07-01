# Билд стадии
FROM golang:1.21-alpine AS builder

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git make

# Создаем рабочую директорию
WORKDIR /app

# Копируем файлы модулей
COPY go.mod go.sum ./

# Скачиваем зависимости
RUN go mod download

# Копируем весь проект
COPY . .

# Собираем приложение с флагами оптимизации
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/weatherbot ./cmd/weatherbot/main.go

# Финальная стадия
FROM alpine:3.19

# Устанавливаем tzdata для работы с временными зонами
RUN apk add --no-cache tzdata ca-certificates

# Копируем бинарник из стадии builder
COPY --from=builder /bin/weatherbot /bin/weatherbot

# Создаем директорию для конфигов
RUN mkdir -p /etc/weatherbot
COPY config /etc/weatherbot/config

# Рабочая директория
WORKDIR /app

# Точка входа
CMD ["/bin/weatherbot"]