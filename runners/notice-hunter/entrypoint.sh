#!/bin/sh
# Notice hunter — Gate 3c: hit matrix URLs and fail on theme path Notice/Warning/Deprecated in debug.log.
set -eu

GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
TARGET="${LAB_TARGET_URL:-http://wordpress}"
THEME_SLUG="${LAB_THEME_SLUG:-}"
LOG="${LAB_DEBUG_LOG:-/var/www/html/wp-content/debug.log}"
URLS_JSON="${LAB_MATRIX_URLS:-[]}"

export WORDPRESS_DB_HOST="${WORDPRESS_DB_HOST:-db:3306}"
export WORDPRESS_DB_USER="${WORDPRESS_DB_USER:-wp}"
export WORDPRESS_DB_PASSWORD="${WORDPRESS_DB_PASSWORD:-wp}"
export WORDPRESS_DB_NAME="${WORDPRESS_DB_NAME:-wordpress}"

cd /var/www/html 2>/dev/null || true

if [ -z "$THEME_SLUG" ] && command -v wp >/dev/null 2>&1; then
  THEME_SLUG="$(wp theme list --status=active --field=name --allow-root 2>/dev/null || true)"
fi
THEME_SLUG="${THEME_SLUG:-latte}"
THEME_NEEDLE="themes/${THEME_SLUG}"

mkdir -p "$(dirname "$LOG")" 2>/dev/null || true
touch "$LOG" 2>/dev/null || true
BASE_SIZE="$(wc -c < "$LOG" 2>/dev/null | tr -d ' ' || echo 0)"

URLS="$(URLS_JSON="$URLS_JSON" TARGET="$TARGET" php -r '
$raw = getenv("URLS_JSON") ?: "[]";
$arr = json_decode($raw, true);
if (!is_array($arr) || count($arr) === 0) {
  $b = rtrim(getenv("TARGET") ?: "http://wordpress", "/");
  $arr = [$b."/", $b."/?p=1", $b."/?page_id=2", $b."/?s=hello", $b."/?p=999999&lab-404=1"];
}
echo implode("\n", $arr);
')"

echo "$URLS" | while IFS= read -r u; do
  [ -n "$u" ] || continue
  curl -fsS -o /dev/null -L --max-time 15 "$u" >/dev/null 2>&1 || true
done

FINDINGS_JSON="$(BASE_SIZE="$BASE_SIZE" LOG="$LOG" THEME_NEEDLE="$THEME_NEEDLE" GATE="$GATE" CHECK="$CHECK" TARGET="$TARGET" php -r '
$log = getenv("LOG");
$base = (int) getenv("BASE_SIZE");
$needle = getenv("THEME_NEEDLE");
$gate = getenv("GATE");
$check = getenv("CHECK");
$target = getenv("TARGET");
$findings = [];
$data = "";
if (is_file($log)) {
  $fh = fopen($log, "rb");
  if ($fh) {
    if ($base > 0) { fseek($fh, $base); }
    $data = stream_get_contents($fh) ?: "";
    fclose($fh);
  }
}
$lines = preg_split("/\r\n|\n|\r/", $data) ?: [];
$hits = 0;
foreach ($lines as $line) {
  if ($line === "") { continue; }
  if (!preg_match("/\b(PHP )?(Notice|Warning|Deprecated|Fatal error)\b/i", $line, $m)) { continue; }
  if ($needle !== "" && stripos($line, $needle) === false) { continue; }
  $hits++;
  $sev = "medium";
  $kind = strtolower($m[2] ?? "notice");
  if (str_contains($kind, "fatal")) { $sev = "high"; }
  if (str_contains($kind, "deprecated")) { $sev = "low"; }
  $findings[] = [
    "code" => "org.notice.found",
    "gate" => $gate,
    "check" => $check,
    "severity" => $sev,
    "message" => substr($line, 0, 400),
    "target" => $target,
    "evidence" => ["kind" => $kind, "theme" => $needle],
  ];
  if ($hits >= 25) { break; }
}
if ($hits === 0) {
  $findings[] = [
    "code" => "org.notice.ok",
    "gate" => $gate,
    "check" => $check,
    "severity" => "info",
    "message" => "No theme Notice/Warning/Deprecated in debug.log for matrix URLs",
    "target" => $target,
    "evidence" => ["logBytes" => (string) strlen($data), "theme" => $needle],
  ];
}
$findings[] = [
  "code" => "org.notice.summary",
  "gate" => $gate,
  "check" => $check,
  "severity" => "info",
  "message" => sprintf("Notice hunter: %d theme issue(s)", $hits),
  "target" => $target,
  "evidence" => ["hits" => (string) $hits],
];
echo json_encode(["findings" => $findings], JSON_UNESCAPED_SLASHES);
')"

printf '%s\n' "$FINDINGS_JSON"
