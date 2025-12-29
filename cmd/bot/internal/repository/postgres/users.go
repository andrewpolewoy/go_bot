package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/andrewpolewoy/go_bot/cmd/bot/internal/repository"
)

var ErrNotFound = errors.New("not found")

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func normalizeLogin(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func (r *UserRepo) SaveBinding(binding repository.UserBinding) error {
	login := normalizeLogin(binding.GitHubLogin)
	if login == "" {
		return errors.New("github login is empty")
	}

	const q = `
INSERT INTO user_bindings (telegram_id, github_login)
VALUES ($1, $2)
ON CONFLICT (telegram_id) DO UPDATE SET github_login = EXCLUDED.github_login;
`
	_, err := r.pool.Exec(context.Background(), q, binding.TelegramID, login)
	return err
}

func (r *UserRepo) GetByTelegramID(tgID int64) (*repository.UserBinding, error) {
	const q = `
SELECT telegram_id, github_login
FROM user_bindings
WHERE telegram_id = $1;
`
	var b repository.UserBinding
	err := r.pool.QueryRow(context.Background(), q, tgID).Scan(&b.TelegramID, &b.GitHubLogin)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get by telegram id: %w", err)
	}
	return &b, nil
}

func (r *UserRepo) GetByGitHubLogin(login string) ([]repository.UserBinding, error) {
	login = normalizeLogin(login)

	const q = `
SELECT telegram_id, github_login
FROM user_bindings
WHERE github_login = $1;
`
	rows, err := r.pool.Query(context.Background(), q, login)
	if err != nil {
		return nil, fmt.Errorf("get by github login: %w", err)
	}
	defer rows.Close()

	var out []repository.UserBinding
	for rows.Next() {
		var b repository.UserBinding
		if err := rows.Scan(&b.TelegramID, &b.GitHubLogin); err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(out) == 0 {
		return nil, nil
	}
	return out, nil
}
