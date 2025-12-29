CREATE TABLE IF NOT EXISTS user_bindings (
  telegram_id  BIGINT PRIMARY KEY,
  github_login TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_user_bindings_github_login ON user_bindings (github_login);
