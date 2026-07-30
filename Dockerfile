#Сборка бинарника
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
COPY vendor ./vendor

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /app/server ./cmd/server

#Легкий финальный образ
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/server /app/server

EXPOSE 8080

CMD ["/app/server"]