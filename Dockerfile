FROM golang:1.22.6 AS builder

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/main.go

FROM alpine:latest

ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 6060

CMD ["./main"]
