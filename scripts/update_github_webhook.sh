#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_TOKEN:?set GITHUB_TOKEN (PAT with repo / write:repo_hook)}"
: "${GITHUB_OWNER:?set GITHUB_OWNER (e.g. andrewpolewoy)}"
: "${GITHUB_REPO:?set GITHUB_REPO (e.g. go_bot)}"
: "${GITHUB_HOOK_ID:?set GITHUB_HOOK_ID (number)}"
: "${PUBLIC_PATH:?set PUBLIC_PATH (e.g. /api/v1/github/webhook)}"

LOG_FILE="${TUNNEL_LOG_FILE:-/tmp/cloudflared-go-bot.log}"

# Берём последний https://*.trycloudflare.com из лога cloudflared
PUBLIC_BASE_URL="$(grep -Eo 'https://[a-zA-Z0-9-]+\.trycloudflare\.com' "${LOG_FILE}" | tail -n 1 || true)"
if [[ -z "${PUBLIC_BASE_URL}" ]]; then
  echo "Could not find trycloudflare URL in log: ${LOG_FILE}" >&2
  exit 1
fi

NEW_URL="${PUBLIC_BASE_URL}${PUBLIC_PATH}"
echo "Updating GitHub webhook to: ${NEW_URL}"

curl -sS -L \
  -X PATCH \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  "https://api.github.com/repos/${GITHUB_OWNER}/${GITHUB_REPO}/hooks/${GITHUB_HOOK_ID}/config" \
  -d "{\"content_type\":\"json\",\"url\":\"${NEW_URL}\"}" \
  >/dev/null

echo "OK"
