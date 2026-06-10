FROM golang:1.23-alpine AS builder

WORKDIR /app

# Копируем go.mod и go.sum
COPY go.mod go.sum* ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o bank-api ./cmd/main.go

# Этап выполнения
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Копируем бинарник и миграции
COPY --from=builder /app/bank-api .
COPY --from=builder /app/migrations ./migrations

EXPOSE 8080

CMD ["./bank-api"]