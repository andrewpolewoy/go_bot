package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository"
)

type TelegramSender interface {
	SendMessage(ctx context.Context, chatID int64, text string) error
}

type Notifier struct {
	users  repository.UserRepository
	sender TelegramSender
}

func NewNotifier(users repository.UserRepository, sender TelegramSender) *Notifier {
	return &Notifier{users: users, sender: sender}
}

func (s *Notifier) SetGitHubLogin(ctx context.Context, tgID int64, login string) error {
	login = strings.TrimSpace(login)
	if login == "" {
		return fmt.Errorf("empty login")
	}
	return s.users.SaveBinding(repository.UserBinding{
		TelegramID:  tgID,
		GitHubLogin: login,
	})
}

func (s *Notifier) GetMe(ctx context.Context, tgID int64) (string, error) {
	b, err := s.users.GetByTelegramID(tgID)
	if err != nil {
		return "", err
	}
	return b.GitHubLogin, nil
}

func (s *Notifier) NotifyAssignee(ctx context.Context, assigneeLogin, msg string) error {
	bindings, err := s.users.GetByGitHubLogin(assigneeLogin)
	if err != nil {
		return err
	}

	for _, b := range bindings {
		if err := s.sender.SendMessage(ctx, b.TelegramID, msg); err != nil {
			return fmt.Errorf("send telegram message to %d: %w", b.TelegramID, err)
		}
	}
	return nil
}
