package app

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	log    *Logger
	cfg    *Config
	server *http.Server
}

func Start() error {
	a := &App{}
	if err := a.Bootstrap(); err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runErr := make(chan error, 1)
	go func() { runErr <- a.Run() }()

	select {
	case <-ctx.Done():
	case err := <-runErr:
		_ = a.Shutdown(10 * time.Second)
		return err
	}

	if err := a.Shutdown(10 * time.Second); err != nil {
		return err
	}
	return nil
}
