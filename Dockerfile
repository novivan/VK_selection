FROM golang:1.20-alpine

WORKDIR /app

# Устанавливаем необходимые пакеты, включая pkgconfig и openssl-dev
RUN apk add --no-cache git gcc musl-dev pkgconfig openssl-dev

# Копируем всё сразу
COPY . .

# Очищаем go.sum и создаем его заново с полной установкой всех зависимостей
RUN rm -f go.sum && \
    go clean -modcache && \
    GOPRIVATE=github.com/tarantool/* GO111MODULE=on go mod tidy && \
    GO111MODULE=on GOPRIVATE=github.com/tarantool/* go build -o pollbot

EXPOSE 8080
CMD ["./pollbot"]
