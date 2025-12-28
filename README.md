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

Runtime:
- `CRNB_SERVER_PORT` (default: 8080)
- `CRNB_SERVER_PUBLIC_URL` (used to set Telegram webhook URL, if enabled)

## Run locally (Go)
```
export CRNB_TELEGRAM_BOT_TOKEN="..."
export CRNB_GITHUB_SECRET="..."
go run ./cmd/bot '
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
crnbot:local
```

## Tests & lint
```
go test ./...
golangci-lint run ./...
```
​