package postgres

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository"
)

func TestUserRepo_SaveAndGetByTelegramID(t *testing.T) {
	pool := newTestPool(t)
	repo := NewUserRepo(pool)

	tgID := time.Now().UnixNano()

	err := repo.SaveBinding(repository.UserBinding{
		TelegramID:  tgID,
		GitHubLogin: "AndrewPolewoy",
	})
	if err != nil {
		t.Fatalf("SaveBinding: %v", err)
	}

	got, err := repo.GetByTelegramID(tgID)
	if err != nil {
		t.Fatalf("GetByTelegramID: %v", err)
	}
	if got == nil {
		t.Fatalf("GetByTelegramID: got nil")
	}
	if got.TelegramID != tgID {
		t.Fatalf("TelegramID: want %d, got %d", tgID, got.TelegramID)
	}
	if got.GitHubLogin != "andrewpolewoy" { // т.к. normalizeLogin делает strings.ToLower
		t.Fatalf("GitHubLogin: want %q, got %q", "andrewpolewoy", got.GitHubLogin)
	}
}

func TestUserRepo_GetByTelegramID_NotFound(t *testing.T) {
	pool := newTestPool(t)
	repo := NewUserRepo(pool)

	tgID := time.Now().UnixNano()

	_, err := repo.GetByTelegramID(tgID)
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got: %v", err)
	}
}

func TestUserRepo_GetByGitHubLogin(t *testing.T) {
	pool := newTestPool(t)
	repo := NewUserRepo(pool)

	suffix := time.Now().UnixNano()
	login := fmt.Sprintf("user_%d", suffix)

	// Два разных tgID → один github login
	_ = repo.SaveBinding(repository.UserBinding{TelegramID: suffix + 1, GitHubLogin: login})
	_ = repo.SaveBinding(repository.UserBinding{TelegramID: suffix + 2, GitHubLogin: login})

	got, err := repo.GetByGitHubLogin(login)
	if err != nil {
		t.Fatalf("GetByGitHubLogin: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 bindings, got %d", len(got))
	}
}
