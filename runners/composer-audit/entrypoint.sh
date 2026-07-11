#!/bin/sh
set -eu
# Audit Composer deps from theme zip (LAB_THEME_ZIP) or mounted composer project.
GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
ZIP="${LAB_THEME_ZIP:-}"
WORKDIR=""

emit() {
  printf '%s\n' "$1"
}

if [ -n "$ZIP" ] && [ -f "$ZIP" ]; then
  WORKDIR="$(mktemp -d)"
  unzip -q "$ZIP" -d "$WORKDIR"
  # Prefer theme-root composer.lock (first shallowest lock under zip root)
  LOCK="$(find "$WORKDIR" -name composer.lock -not -path '*/vendor/*' | awk '{print length, $0}' | sort -n | head -1 | cut -d' ' -f2-)"
  if [ -z "$LOCK" ]; then
    emit "{\"findings\":[{\"code\":\"sec.composer.lock_missing\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"info\",\"message\":\"no composer.lock in theme zip (outside vendor)\",\"target\":\"$ZIP\"}]}"
    exit 0
  fi
  CD="$(dirname "$LOCK")"
else
  emit "{\"findings\":[{\"code\":\"sec.composer.zip_missing\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"medium\",\"message\":\"LAB_THEME_ZIP not mounted\",\"target\":\"\"}]}"
  exit 0
fi

cd "$CD"
# composer audit writes JSON to stdout; non-zero exit when advisories exist
set +e
RAW="$(composer audit --format=json --no-interaction 2>/dev/null)"
RC=$?
set -e
if [ -z "$RAW" ]; then
  emit "{\"findings\":[{\"code\":\"sec.composer.exec_failed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"composer audit produced no JSON (rc=$RC)\",\"target\":\"$CD\"}]}"
  exit 0
fi

export RAW GATE CHECK CD
php -r '
$raw = getenv("RAW");
$gate = getenv("GATE");
$check = getenv("CHECK");
$target = getenv("CD");
$findings = [];
$data = json_decode($raw, true);
if (!is_array($data)) {
  echo json_encode(["findings"=>[["code"=>"sec.composer.exec_failed","gate"=>$gate,"check"=>$check,"severity"=>"high","message"=>"invalid composer audit JSON","target"=>$target]]]);
  exit(0);
}
$advisories = $data["advisories"] ?? [];
$count = 0;
if (is_array($advisories)) {
  foreach ($advisories as $pkg => $list) {
    if (!is_array($list)) continue;
    foreach ($list as $adv) {
      if (!is_array($adv)) continue;
      $title = $adv["title"] ?? $adv["advisoryId"] ?? "advisory";
      $sev = strtolower((string)($adv["severity"] ?? "high"));
      if (!in_array($sev, ["critical","high","medium","low","info"], true)) $sev = "high";
      $findings[] = [
        "code" => "sec.composer.advisory",
        "gate" => $gate,
        "check" => $check,
        "severity" => $sev,
        "message" => $pkg . ": " . substr((string)$title, 0, 200),
        "target" => $target,
        "evidence" => [
          "package" => (string)$pkg,
          "advisoryId" => (string)($adv["advisoryId"] ?? ""),
        ],
      ];
      $count++;
      if ($count >= 30) break 2;
    }
  }
}
$abandoned = $data["abandoned"] ?? [];
if (is_array($abandoned) && count($abandoned) > 0) {
  $findings[] = [
    "code" => "sec.composer.abandoned",
    "gate" => $gate,
    "check" => $check,
    "severity" => "low",
    "message" => "abandoned packages: " . implode(", ", array_slice(array_keys($abandoned), 0, 10)),
    "target" => $target,
  ];
}
$findings[] = [
  "code" => "sec.composer.ok",
  "gate" => $gate,
  "check" => $check,
  "severity" => "info",
  "message" => $count === 0 ? "composer audit: no advisories" : "composer audit completed with advisories",
  "target" => $target,
  "evidence" => ["advisories" => (string)$count],
];
// rename ok when clean
if ($count === 0) {
  $findings[count($findings)-1]["code"] = "sec.composer.ok";
} else {
  $findings[count($findings)-1]["code"] = "sec.composer.completed";
}
echo json_encode(["findings" => $findings], JSON_UNESCAPED_SLASHES);
'
