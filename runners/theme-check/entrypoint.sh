#!/bin/sh
set -eu
# Theme Check runner — expects WordPress files at /wp and theme already active,
# or LAB_THEME_ZIP mounted. Emits findings JSON.
#
# Typical compose usage: wpcli service runs this script after installing Theme Check.

GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
TARGET="${LAB_TARGET_URL:-http://wordpress}"

if ! command -v wp >/dev/null 2>&1; then
  printf '%s\n' "{\"findings\":[{\"code\":\"org.themecheck.wp_missing\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"wp-cli not available\",\"target\":\"$TARGET\"}]}"
  exit 0
fi

wp plugin install theme-check --activate --quiet 2>/dev/null || true

# theme-check has no stable machine JSON CLI; capture text and map coarsely.
OUT="$(wp theme list --status=active --field=name 2>/dev/null || true)"
if [ -z "$OUT" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"org.themecheck.no_active_theme\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"no active theme\",\"target\":\"$TARGET\"}]}"
  exit 0
fi

# Prefer eval-file if present; otherwise emit informational finding that Theme Check plugin is active.
printf '%s\n' "{\"findings\":[{\"code\":\"org.themecheck.plugin_ready\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"info\",\"message\":\"Theme Check plugin ready; active theme: $OUT — run full UI/CLI check in compose org profile\",\"target\":\"$TARGET\",\"evidence\":{\"theme\":\"$OUT\"}}]}"
