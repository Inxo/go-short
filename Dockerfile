# Используем образ Golang для сборки
FROM golang:alpine AS builder
RUN apk --no-cache add gcc g++ make git
# Устанавливаем переменную окружения для работы внутри контейнера
ENV GO111MODULE=on \
    CGO_ENABLED=1 \
    GOOS=linux \
    GOARCH=amd64

# Копируем исходный код в Docker-контейнер
WORKDIR /app
COPY . .

# Собираем приложение
RUN go mod tidy
RUN go mod vendor
RUN make all

# Второй этап: Создаем минимальный образ для запуска
FROM alpine
RUN apk --no-cache add ca-certificates
WORKDIR /app
# Копируем бинарный файл из предыдущего образа
COPY --from=builder /app/build /app/bin
COPY --from=builder /app/templates /app/templates
COPY  /data/database.db /app/data/database.db
EXPOSE 8080
# Определяем команду для запуска контейнера
ENTRYPOINT /app/bin/my-url-shortener
