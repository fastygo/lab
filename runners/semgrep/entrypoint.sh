#!/bin/sh
set -eu
GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
ZIP="${LAB_THEME_ZIP:-}"

if [ -z "$ZIP" ] || [ ! -f "$ZIP" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"sec.semgrep.zip_missing\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"medium\",\"message\":\"LAB_THEME_ZIP not mounted\",\"target\":\"\"}]}"
  exit 0
fi

WORKDIR="$(mktemp -d)"
unzip -q "$ZIP" -d "$WORKDIR"
THEME_ROOT="$(find "$WORKDIR" -mindepth 1 -maxdepth 1 -type d | head -1)"
if [ -z "$THEME_ROOT" ]; then
  THEME_ROOT="$WORKDIR"
fi

set +e
RAW="$(semgrep --config /rules/lab-theme-sec.yml --json --quiet \
  --exclude=vendor --exclude=node_modules --exclude='~cache' \
  "$THEME_ROOT" 2>/dev/null)"
RC=$?
set -e

if [ -z "$RAW" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"sec.semgrep.exec_failed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"semgrep produced no JSON (rc=$RC)\",\"target\":\"$THEME_ROOT\"}]}"
  exit 0
fi

export RAW GATE CHECK THEME_ROOT RC
python3 - <<'PY'
import json, os
raw = os.environ["RAW"]
gate = os.environ["GATE"]
check = os.environ["CHECK"]
target = os.environ["THEME_ROOT"]
findings = []
try:
    data = json.loads(raw)
except Exception as e:
    print(json.dumps({"findings":[{"code":"sec.semgrep.exec_failed","gate":gate,"check":check,"severity":"high","message":f"invalid semgrep JSON: {e}","target":target}]}))
    raise SystemExit(0)

n = 0
for r in data.get("results") or []:
    check_id = r.get("check_id") or "semgrep"
    path = r.get("path") or ""
    if path.startswith(target):
        path = path[len(target):].lstrip("/")
    msg = (r.get("extra") or {}).get("message") or check_id
    sev_raw = ((r.get("extra") or {}).get("severity") or "WARNING").upper()
    sev = "high" if sev_raw in ("ERROR", "CRITICAL") else "medium"
    findings.append({
        "code": "sec.semgrep.finding",
        "gate": gate, "check": check,
        "severity": sev,
        "message": f"{path}: {msg}"[:240],
        "target": target,
        "evidence": {"rule": str(check_id)},
    })
    n += 1
    if n >= 40:
        break

findings.append({
    "code": "sec.semgrep.ok" if n == 0 else "sec.semgrep.completed",
    "gate": gate, "check": check,
    "severity": "info",
    "message": "semgrep: no findings" if n == 0 else "semgrep completed with findings",
    "target": target,
    "evidence": {"issues": str(n), "rc": os.environ.get("RC", "")},
})
print(json.dumps({"findings": findings}))
PY
