package app

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	log    *Logger      // см. ниже
	cfg    *Config      // обёртка вокруг твоего config.Config
	server *http.Server // HTTP сервер с Telegram/GitHub handlers
}

func Start() {
	a := &App{}
	if err := a.Bootstrap(); err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go a.Run()

	<-ctx.Done()
	if err := a.Shutdown(10 * time.Second); err != nil {
		a.log.Error("shutdown error", "err", err)
	} else {
		a.log.Info("bot gracefully stopped")
	}
}
