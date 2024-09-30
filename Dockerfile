# Используем официальный образ Go для сборки
FROM golang:1.21 as builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем файлы проекта
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# Выполнить go mod tidy для обновления go.mod и go.sum
RUN go mod tidy

# Сборка бинарного файла
RUN CGO_ENABLED=0 GOOS=linux go build -o /task-tracker cmd/main.go

# Создаем финальный сжатый образ
FROM alpine:latest

# Устанавливаем зависимости
RUN apk --no-cache add ca-certificates

# Копируем бинарный файл из предыдущего шага
COPY --from=builder /task-tracker /task-tracker

# Запуск приложения
CMD ["/task-tracker"]