package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// TODO: загрузить конфиг
	// TODO: инициализировать логгер
	// TODO: собрать репозиторий, сервис, delivery-слои
	// TODO: запустить HTTP-сервер (GitHub webhook) + Telegram-bot обработчик

	log.Println("code-review-notifier-bot started (stub)")

	<-ctx.Done()
	log.Println("shutting down...")
}
