package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository"
	pgrepo "github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository/postgres"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	appcfg "github.com/andrewpolewoy/go_bot/cmd/bot/internal/config"
	httpdelivery "github.com/andrewpolewoy/go_bot/cmd/bot/internal/delivery/http"
	tgdelivery "github.com/andrewpolewoy/go_bot/cmd/bot/internal/delivery/telegram"
	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/logger"
	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository/memory"
	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/service"
)

type Config struct {
	Raw appcfg.Config
}

func (a *App) Bootstrap() error {
	rawCfg, err := appcfg.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	a.cfg = &Config{Raw: rawCfg}

	// Initialize logger with level from config
	logLevel := logger.Level(rawCfg.Log.Level)
	if logLevel == "" {
		logLevel = logger.LevelInfo
	}
	a.log = logger.New(logLevel)

	if rawCfg.Telegram.BotToken == "" {
		return fmt.Errorf("telegram bot token is empty")
	}

	a.log.Info("bootstrapping bot")

	// dependencies
	var repo repository.UserRepository

	if rawCfg.DB.DSN != "" {
		pool, err := pgxpool.New(context.Background(), rawCfg.DB.DSN)
		if err != nil {
			return fmt.Errorf("connect postgres: %w", err)
		}

		a.db = pool
		repo = pgrepo.NewUserRepo(pool)
		a.log.Info("using postgres repository")
	} else {
		repo = memory.NewUserRepo()
		a.log.Info("using memory repository")
	}

	bot, err := tgbotapi.NewBotAPI(rawCfg.Telegram.BotToken)
	if err != nil {
		return fmt.Errorf("create telegram bot: %w", err)
	}
	bot.Debug = false

	sender := tgdelivery.NewSender(bot)
	svc := service.NewNotifier(repo, sender, a.log)
	tgHandler := tgdelivery.NewHandler(svc, bot, a.log)

	ghHandler := httpdelivery.NewHandler(svc, rawCfg.Github.Secret, a.log)

	// webhook setup (Telegram)
	if rawCfg.Server.PublicURL != "" {
		webhookURL := rawCfg.Server.PublicURL + rawCfg.Server.TelegramWebhookPath
		wh, _ := tgbotapi.NewWebhook(webhookURL)
		if _, err := bot.Request(wh); err != nil {
			return fmt.Errorf("set telegram webhook: %w", err)
		}
		a.log.Info("telegram webhook set", "url", webhookURL)
	}

	// HTTP mux
	mux := http.NewServeMux()
	mux.Handle(rawCfg.Server.GithubWebhookPath, http.HandlerFunc(ghHandler.GitHubWebhook))
	mux.HandleFunc(rawCfg.Server.TelegramWebhookPath, func(w http.ResponseWriter, r *http.Request) {
		update, err := bot.HandleUpdate(r)
		if err != nil {
			a.log.Error("telegram handle update error", "err", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		tgHandler.HandleUpdate(*update)
		w.WriteHeader(http.StatusOK)
	})

	addr := net.JoinHostPort("", fmt.Sprintf("%d", rawCfg.Server.Port))

	a.server = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	a.log.Info("bot bootstrapped", "addr", addr)
	return nil
}
