FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bot ./cmd/bot

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=builder /app/bot /app/bot
COPY config/config.yml.example /app/config/config.yml

ENV CRNB_SERVER_PORT=8080

EXPOSE 8080

ENTRYPOINT ["/app/bot"]
