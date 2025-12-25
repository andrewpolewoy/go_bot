package memory

import (
	"errors"
	"strings"
	"sync"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository"
)

var ErrNotFound = errors.New("not found")

type UserRepo struct {
	mu      sync.RWMutex
	byTG    map[int64]repository.UserBinding
	byLogin map[string]map[int64]struct{} // login -> set(tgID)
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		byTG:    make(map[int64]repository.UserBinding),
		byLogin: make(map[string]map[int64]struct{}),
	}
}

func normalizeLogin(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (r *UserRepo) SaveBinding(binding repository.UserBinding) error {
	login := normalizeLogin(binding.GitHubLogin)
	if login == "" {
		return errors.New("github login is empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// если у tgID уже была старая привязка — почистим индекс byLogin
	if old, ok := r.byTG[binding.TelegramID]; ok {
		oldLogin := normalizeLogin(old.GitHubLogin)
		if set, ok := r.byLogin[oldLogin]; ok {
			delete(set, binding.TelegramID)
			if len(set) == 0 {
				delete(r.byLogin, oldLogin)
			}
		}
	}

	r.byTG[binding.TelegramID] = repository.UserBinding{
		TelegramID:  binding.TelegramID,
		GitHubLogin: login,
	}

	set, ok := r.byLogin[login]
	if !ok {
		set = make(map[int64]struct{})
		r.byLogin[login] = set
	}
	set[binding.TelegramID] = struct{}{}
	return nil
}

func (r *UserRepo) GetByTelegramID(tgID int64) (*repository.UserBinding, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	b, ok := r.byTG[tgID]
	if !ok {
		return nil, ErrNotFound
	}
	cp := b
	return &cp, nil
}

func (r *UserRepo) GetByGitHubLogin(login string) ([]repository.UserBinding, error) {
	login = normalizeLogin(login)

	r.mu.RLock()
	defer r.mu.RUnlock()

	set, ok := r.byLogin[login]
	if !ok || len(set) == 0 {
		return nil, nil
	}

	out := make([]repository.UserBinding, 0, len(set))
	for tgID := range set {
		out = append(out, r.byTG[tgID])
	}
	return out, nil
}
