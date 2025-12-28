package telegram

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/service"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Sender struct {
	bot *tgbotapi.BotAPI
}

func NewSender(bot *tgbotapi.BotAPI) *Sender {
	return &Sender{bot: bot}
}

func (s *Sender) SendMessage(ctx context.Context, chatID int64, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	_, err := s.bot.Send(msg)
	return err
}

type Handler struct {
	svc *service.Notifier
	bot *tgbotapi.BotAPI
}

func NewHandler(svc *service.Notifier, bot *tgbotapi.BotAPI) *Handler {
	return &Handler{svc: svc, bot: bot}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update) {
	if update.Message == nil || update.Message.Text == "" {
		return
	}

	ctx := context.TODO()
	chatID := update.Message.Chat.ID
	text := strings.TrimSpace(update.Message.Text)

	log.Printf("received message: %s from %d", text, chatID)

	var reply string

	switch {
	case text == "/start":
		reply = "Привет! Команды: /setgithub <login>, /me"

	case strings.HasPrefix(text, "/setgithub"):
		parts := strings.Fields(text)
		if len(parts) != 2 {
			reply = "Использование: /setgithub <github_login>"
		} else {
			login := parts[1]
			if err := h.svc.SetGitHubLogin(ctx, chatID, login); err != nil {
				reply = fmt.Sprintf("Ошибка: %v", err)
			} else {
				reply = "Ок, сохранил."
			}
		}

	case text == "/me":
		login, err := h.svc.GetMe(ctx, chatID)
		if err != nil {
			reply = "Пока не задан github login. Используй /setgithub <login>."
		} else {
			reply = "Твой GitHub login: " + login
		}

	default:
		return
	}

	msg := tgbotapi.NewMessage(chatID, reply)
	_, _ = h.bot.Send(msg)
}
