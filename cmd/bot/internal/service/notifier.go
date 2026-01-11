package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/logger"
	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository"
)

type TelegramSender interface {
	SendMessage(chatID int64, text string) error
}

type Notifier struct {
	users  repository.UserRepository
	sender TelegramSender
	logger logger.Logger
}

func NewNotifier(users repository.UserRepository, sender TelegramSender, log logger.Logger) *Notifier {
	if log == nil {
		log = logger.NewDefault()
	}
	return &Notifier{users: users, sender: sender, logger: log}
}

func (s *Notifier) SetGitHubLogin(ctx context.Context, tgID int64, login string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return fmt.Errorf("empty login")
	}
	return s.users.SaveBinding(ctx, repository.UserBinding{
		TelegramID:  tgID,
		GitHubLogin: login,
	})
}

func (s *Notifier) GetMe(ctx context.Context, tgID int64) (string, error) {
	b, err := s.users.GetByTelegramID(ctx, tgID)
	if err != nil {
		return "", err
	}
	return b.GitHubLogin, nil
}

func (s *Notifier) NotifyAssignee(ctx context.Context, assigneeLogin, msg string) error {
	bindings, err := s.users.GetByGitHubLogin(ctx, assigneeLogin)
	if err != nil {
		return fmt.Errorf("get bindings for github login %q: %w", assigneeLogin, err)
	}

	if len(bindings) == 0 {
		s.logger.Debug("no telegram bindings found for github login", "login", assigneeLogin)
		return nil
	}

	var lastErr error
	for _, b := range bindings {
		if err := s.sender.SendMessage(b.TelegramID, msg); err != nil {
			// Log error but continue sending to other users
			s.logger.Error("failed to send telegram message", "err", err, "chat_id", b.TelegramID, "github_login", assigneeLogin)
			lastErr = err
			// Continue to next binding
		}
	}

	// Return last error if all sends failed, but don't fail if at least one succeeded
	if lastErr != nil && len(bindings) == 1 {
		return fmt.Errorf("send telegram message to %d: %w", bindings[0].TelegramID, lastErr)
	}

	return nil
}
