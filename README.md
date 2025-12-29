# Code Review Notifier Bot

Telegram bot + HTTP service that sends **personal** notifications about GitHub Pull Request review workflow.

## Tech stack
- Go 1.24
- net/http (HTTP server)
- Telegram Bot API (`go-telegram-bot-api/v5`)
- GitHub Webhooks + HMAC SHA-256 signature validation (`X-Hub-Signature-256`)
- Docker (multi-stage build, distroless runtime)
- GitHub Actions CI (tests + golangci-lint)

## Features
- Telegram commands:
  - `/setgithub <github_login>` — bind Telegram `chat_id` to GitHub login
  - `/me` — show saved GitHub login
- GitHub webhook endpoint:
  - validates webhook signature (HMAC secret)
  - processes events:
    - `pull_request` (`action=assigned`)
    - `pull_request_review` (`action=submitted`)
    - `pull_request_review_comment` (`action=created`)

## Architecture (layers)
- `delivery/http` — GitHub webhook handler
- `delivery/telegram` — Telegram handler + sender
- `service` — business logic (bind user, notify)
- `repository` — storage abstraction (`memory` implementation)

## Configuration (env)
Required:
- `CRNB_TELEGRAM_BOT_TOKEN`
- `CRNB_GITHUB_SECRET`

Optional:
- `CRNB_DB_DSN` — if empty, uses in-memory repository; if set, uses PostgreSQL repository.
  Example: `postgres://crnbot:crnbot@localhost:5432/crnbot?sslmode=disable`

Runtime:
- `CRNB_SERVER_PORT` (default: 8080)
- `CRNB_SERVER_PUBLIC_URL` (used to set Telegram webhook URL, if enabled)


## Run locally (Go)
Start PostgreSQL (optional):
```
docker compose up -d db
```
Run migrations (if DB is used):
```
make migrate-up
```
Run app:
```
export CRNB_TELEGRAM_BOT_TOKEN="..."
export CRNB_GITHUB_SECRET="..."
export CRNB_DB_DSN="postgres://crnbot:crnbot@localhost:5432/crnbot?sslmode=disable" # optional
go run ./cmd/bot
```
## Migrations

This project uses `golang-migrate` CLI.
- Migration files live in `./migrations` (`*.up.sql` / `*.down.sql`).
- Current DB schema version is tracked in DB table `schema_migrations`.

Common commands:
```
make migrate-up
make migrate-down
make migrate-version
make migrate-create NAME=add_table_name
```

## Expose localhost to GitHub (Cloudflare quick tunnel)
Cloudflare quick tunnel URL changes on every start.
Use scripts to update GitHub webhook automatically.

Terminal 1 (start tunnel and save logs):
```
PORT=8080 ./scripts/dev-tunnel.sh
```

Terminal 2 (update GitHub webhook to new public URL):
```
export GITHUB_TOKEN="..." # PAT with repo hooks permissions
export GITHUB_OWNER="andrewpolewoy"
export GITHUB_REPO="go_bot"
export GITHUB_HOOK_ID="123456"
export PUBLIC_PATH="/api/v1/github/webhook"
bash ./scripts/update_github_webhook.sh
```

## Run with Docker
Build image:
```
docker build -t crnbot:local .
```

Run container:
```
docker run --rm -p 8080:8080
-e CRNB_TELEGRAM_BOT_TOKEN="..."
-e CRNB_GITHUB_SECRET="..."
-e CRNB_DB_DSN="postgres://crnbot:crnbot@host.docker.internal:5432/crnbot?sslmode=disable"
crnbot:local
```

## Tests & lint
```
go test ./...
golangci-lint run ./...
```
​