# Этап сборки: используем официальный образ Golang
FROM golang:1.20 AS builder

WORKDIR /app

# Копируем файлы зависимостей и загружаем модули
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код и компилируем приложение, отключая CGO для статической сборки
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o go-robot ./cmd/server

# Этап исполнения: используем минимальный образ Alpine
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/
COPY --from=builder /app/go-robot .

EXPOSE 8080

CMD ["./go-robot"]
