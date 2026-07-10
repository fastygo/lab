#!/bin/sh
set -eu
# Thin wrapper around wpscanteam/wpscan — maps text output to findings JSON.
TARGET="${LAB_TARGET_URL:-http://wordpress/}"
GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
TOKEN_ARGS=""
if [ -n "${WPSCAN_API_TOKEN:-}" ]; then
  TOKEN_ARGS="--api-token ${WPSCAN_API_TOKEN}"
fi

# shellcheck disable=SC2086
RAW="$(wpscan --url "$TARGET" --enumerate u,t,p --format json $TOKEN_ARGS 2>/dev/null || true)"
if [ -z "$RAW" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"sec.wpscan.exec_failed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"wpscan produced no JSON\",\"target\":\"$TARGET\"}]}"
  exit 0
fi

# Pass through a minimal finding; full CVE mapping can be refined later.
printf '%s\n' "{\"findings\":[{\"code\":\"sec.wpscan.completed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"info\",\"message\":\"wpscan enumeration completed\",\"target\":\"$TARGET\",\"evidence\":{\"bytes\":\"${#RAW}\"}}]}"
