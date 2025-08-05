FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN GOOS=linux GOARCH=amd64 go build -o talkify cmd/server/main.go

FROM alpine:3.18

WORKDIR /app
COPY --from=builder /app/talkify .
COPY migrations ./migrations

RUN chmod +x /app/talkify

EXPOSE 8080

CMD ["./talkify"]