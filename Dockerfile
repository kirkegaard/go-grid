FROM golang:1.22.6 AS builder

ENV GOOS=linux
ENV GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum www ./

RUN go mod download

COPY . .

RUN go build -o main ./cmd/server/

FROM alpine:latest

ENV GRID_PORT=6060
ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

WORKDIR /root/

COPY --from=builder /app/main .
COPY --from=builder /app/www www

EXPOSE 6060

CMD ["./main"]
