#!/bin/sh
set -eu
GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
ZIP="${LAB_THEME_ZIP:-}"

if [ -z "$ZIP" ] || [ ! -f "$ZIP" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"sec.phpcs.zip_missing\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"medium\",\"message\":\"LAB_THEME_ZIP not mounted\",\"target\":\"\"}]}"
  exit 0
fi

WORKDIR="$(mktemp -d)"
unzip -q "$ZIP" -d "$WORKDIR"
# Prefer theme PHP tree (skip vendor bulk — still scanned lightly via --ignore)
THEME_ROOT="$(find "$WORKDIR" -maxdepth 2 -type d -name '*' | awk 'NR==2{print; exit}')"
# First directory under workdir is usually the theme slug
THEME_ROOT="$(find "$WORKDIR" -mindepth 1 -maxdepth 1 -type d | head -1)"
if [ -z "$THEME_ROOT" ]; then
  THEME_ROOT="$WORKDIR"
fi

set +e
REPORT="$(phpcs -q --standard=Security --report=json \
  --extensions=php \
  --ignore=*/vendor/*,*/node_modules/*,*/~cache/* \
  "$THEME_ROOT" 2>/dev/null)"
RC=$?
set -e

if [ -z "$REPORT" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"sec.phpcs.exec_failed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"phpcs produced no JSON (rc=$RC)\",\"target\":\"$THEME_ROOT\"}]}"
  exit 0
fi

export REPORT GATE CHECK THEME_ROOT RC
php -r '
$raw = getenv("REPORT");
$gate = getenv("GATE");
$check = getenv("CHECK");
$target = getenv("THEME_ROOT");
$findings = [];
$data = json_decode($raw, true);
if (!is_array($data)) {
  echo json_encode(["findings"=>[["code"=>"sec.phpcs.exec_failed","gate"=>$gate,"check"=>$check,"severity"=>"high","message"=>"invalid phpcs JSON","target"=>$target]]]);
  exit(0);
}
$files = $data["files"] ?? [];
$n = 0;
foreach ($files as $path => $info) {
  if (!is_array($info)) continue;
  foreach (($info["messages"] ?? []) as $msg) {
    if (!is_array($msg)) continue;
    $type = strtoupper((string)($msg["type"] ?? "WARNING"));
    $sev = $type === "ERROR" ? "high" : "medium";
    $source = (string)($msg["source"] ?? "Security");
    $message = (string)($msg["message"] ?? "phpcs-security finding");
    $line = (string)($msg["line"] ?? "");
    $rel = $path;
    if (str_starts_with($rel, $target)) {
      $rel = ltrim(substr($rel, strlen($target)), "/");
    }
    $findings[] = [
      "code" => "sec.phpcs.security",
      "gate" => $gate,
      "check" => $check,
      "severity" => $sev,
      "message" => substr($rel . ":" . $line . " " . $message, 0, 240),
      "target" => $target,
      "evidence" => ["source" => $source, "type" => $type],
    ];
    $n++;
    if ($n >= 40) break 2;
  }
}
$findings[] = [
  "code" => $n === 0 ? "sec.phpcs.ok" : "sec.phpcs.completed",
  "gate" => $gate,
  "check" => $check,
  "severity" => "info",
  "message" => $n === 0 ? "phpcs-security: no findings" : "phpcs-security completed with findings",
  "target" => $target,
  "evidence" => ["issues" => (string)$n, "rc" => (string)getenv("RC")],
];
echo json_encode(["findings" => $findings], JSON_UNESCAPED_SLASHES);
'
