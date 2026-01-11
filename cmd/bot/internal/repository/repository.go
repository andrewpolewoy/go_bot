package repository

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("not found")

type UserBinding struct {
	TelegramID  int64
	GitHubLogin string
}

type UserRepository interface {
	SaveBinding(ctx context.Context, binding UserBinding) error
	GetByGitHubLogin(ctx context.Context, login string) ([]UserBinding, error)
	GetByTelegramID(ctx context.Context, tgID int64) (*UserBinding, error)
}
