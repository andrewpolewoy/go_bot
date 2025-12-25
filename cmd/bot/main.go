package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	appcfg "github.com/andrewpolewoy/go_bot/cmd/bot/internal/config"
	httpdelivery "github.com/andrewpolewoy/go_bot/cmd/bot/internal/delivery/http"
	tgdelivery "github.com/andrewpolewoy/go_bot/cmd/bot/internal/delivery/telegram"
	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository/memory"
	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/service"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := appcfg.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if cfg.Telegram.BotToken == "" {
		log.Fatalf("telegram bot token is empty (set CRNB_TELEGRAM_BOT_TOKEN)")
	}

	repo := memory.NewUserRepo()

	bot, err := tgbotapi.NewBotAPI(cfg.Telegram.BotToken)
	if err != nil {
		log.Fatalf("create telegram bot: %v", err)
	}
	bot.Debug = false

	sender := tgdelivery.NewSender(bot)
	svc := service.NewNotifier(repo, sender)
	tgHandler := tgdelivery.NewHandler(svc, bot)
	ghHandler := httpdelivery.NewHandler(svc, cfg.Github.Secret)

	// Webhook setup
	if cfg.Server.PublicURL != "" {
		webhookURL := cfg.Server.PublicURL + cfg.Server.TelegramWebhookPath
		wh, _ := tgbotapi.NewWebhook(webhookURL)
		_, err = bot.Request(wh)
		if err != nil {
			log.Fatalf("setWebhook: %v", err)
		}
		log.Printf("telegram webhook set: %s", webhookURL)
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Server.GithubWebhookPath, http.HandlerFunc(ghHandler.GitHubWebhook))

	// Telegram webhook handler
	mux.HandleFunc(cfg.Server.TelegramWebhookPath, func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			log.Printf("handle update error: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tgHandler.HandleUpdate(*update)
		w.WriteHeader(http.StatusOK)
	})

	srv := &http.Server{
		Addr:              net.JoinHostPort("", fmt.Sprintf("%d", cfg.Server.Port)),
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server error: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	log.Println("shutting down...")
}
