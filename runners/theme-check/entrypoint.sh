#!/bin/sh
set -eu
# Theme Check headless runner (Gate 2).
# Env: LAB_TARGET_URL, LAB_GATE_ID, LAB_CHECK_ID, LAB_THEME_ZIP, LAB_CONFIG_JSON
# Emits findings JSON on stdout.

GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
TARGET="${LAB_TARGET_URL:-http://wordpress}"
ZIP="${LAB_THEME_ZIP:-}"
SLUG=""

cd /var/www/html 2>/dev/null || true

findings_error() {
  code="$1"
  sev="$2"
  msg="$3"
  printf '%s\n' "{\"findings\":[{\"code\":\"${code}\",\"gate\":\"${GATE}\",\"check\":\"${CHECK}\",\"severity\":\"${sev}\",\"message\":$(printf '%s' "$msg" | php -r 'echo json_encode(stream_get_contents(STDIN));'),\"target\":\"${TARGET}\"}]}"
}

if ! command -v wp >/dev/null 2>&1; then
  findings_error "org.themecheck.wp_missing" "high" "wp-cli not available"
  exit 0
fi

# Wait for database / WordPress files
i=0
while [ "$i" -lt 60 ]; do
  if wp core is-installed --allow-root 2>/dev/null; then
    break
  fi
  if [ -f wp-config.php ] || [ -f /var/www/html/wp-config.php ]; then
    wp core install \
      --url="${TARGET}" \
      --title="FastyGo Lab" \
      --admin_user=admin \
      --admin_password=admin \
      --admin_email=lab@example.test \
      --skip-email \
      --allow-root 2>/dev/null && break
  fi
  i=$((i + 1))
  sleep 2
done

if ! wp core is-installed --allow-root 2>/dev/null; then
  findings_error "org.themecheck.wp_not_ready" "high" "WordPress not installed/ready for Theme Check"
  exit 0
fi

# Install theme from zip if provided
if [ -n "$ZIP" ] && [ -f "$ZIP" ]; then
  wp theme install "$ZIP" --force --activate --allow-root 2>/dev/null || {
    findings_error "org.themecheck.theme_install_failed" "high" "failed to install theme zip: $ZIP"
    exit 0
  }
elif [ -n "$ZIP" ]; then
  findings_error "org.themecheck.theme_zip_missing" "high" "LAB_THEME_ZIP not found: $ZIP"
  exit 0
fi

SLUG="$(wp theme list --status=active --field=name --allow-root 2>/dev/null || true)"
if [ -z "$SLUG" ]; then
  findings_error "org.themecheck.no_active_theme" "high" "no active theme"
  exit 0
fi

wp plugin install theme-check --activate --force --allow-root 2>/dev/null || true

export LAB_THEME_SLUG="$SLUG"

# Prefer CLI if present; else eval-file headless API
RAW="$(wp theme-check run "$SLUG" --format=json --allow-root 2>/dev/null || true)"
if [ -z "$RAW" ]; then
  RAW="$(wp eval-file /runner/run-check.php --allow-root 2>/dev/null || true)"
fi

if [ -z "$RAW" ]; then
  findings_error "org.themecheck.exec_failed" "high" "Theme Check produced no JSON (plugin CLI and eval-file both empty)"
  exit 0
fi

export LAB_GATE_ID="$GATE" LAB_CHECK_ID="$CHECK" LAB_TARGET_URL="$TARGET" LAB_THEME_SLUG="$SLUG"
printf '%s' "$RAW" | php /runner/to-findings.php
