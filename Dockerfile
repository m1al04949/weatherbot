# Используем легковесный образ Go
FROM golang:1.24-alpine

# Создаем рабочую директорию
WORKDIR /weatherbot

# Копируем зависимости
COPY go.mod go.sum ./
RUN go mod download

# Копируем код
COPY . .

# Собираем приложение
RUN go build -o bot

# Запускаем бота
CMD ["./bot"]