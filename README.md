# Code Review Notifier Bot

Telegram bot that sends personal notifications when GitHub pull request events happen (mainly `pull_request.assigned`).

## Features
- `/setgithub <github_login>` — bind Telegram chat to GitHub login
- `/me` — show saved GitHub login
- GitHub webhook endpoint validates `X-Hub-Signature-256` (HMAC SHA-256)

## Local run
export CRNB_TELEGRAM_BOT_TOKEN="..."
export CRNB_GITHUB_SECRET="..."
export CRNB_SERVER_PUBLIC_URL="https://<your>.trycloudflare.com"
go run ./cmd/bot

## Expose locally (Cloudflare quick tunnel)
PORT=8080 ./scripts/dev-tunnel.sh
bash ./scripts/update_github_webhook.sh

## Test
go test ./...
golangci-lint run ./...