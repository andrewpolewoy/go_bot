#!/usr/bin/env bash
set -euo pipefail

PORT="${PORT:-8080}"
LOCAL_URL="http://localhost:${PORT}"

LOG_FILE="${TUNNEL_LOG_FILE:-/tmp/cloudflared-go-bot.log}"

# Запускаем cloudflared и пишем лог
cloudflared tunnel --url "${LOCAL_URL}" --loglevel info 2>&1 | tee "${LOG_FILE}"
