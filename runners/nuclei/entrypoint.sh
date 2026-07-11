#!/bin/sh
set -eu
# Nuclei WordPress / exposure templates against owned lab target only.
TARGET="${LAB_TARGET_URL:-http://127.0.0.1:8080}"
GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
# Optional: LAB_NUCLEI_TAGS (default wordpress), LAB_NUCLEI_SEVERITY
TAGS="${LAB_NUCLEI_TAGS:-wordpress}"
SEV="${LAB_NUCLEI_SEVERITY:-critical,high,medium}"

# Keep lab runs bounded.
set +e
RAW="$(nuclei -u "$TARGET" -tags "$TAGS" -severity "$SEV" \
  -jsonl -silent -nc -duc \
  -c 10 -rl 50 -timeout 8 -mhe 15 \
  2>/dev/null)"
RC=$?
set -e

if ! command -v python3 >/dev/null 2>&1; then
  printf '%s\n' "{\"findings\":[{\"code\":\"sec.nuclei.completed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"info\",\"message\":\"nuclei finished (no python to map JSONL, rc=$RC)\",\"target\":\"$TARGET\",\"evidence\":{\"bytes\":\"${#RAW}\"}}]}"
  exit 0
fi

export RAW GATE CHECK TARGET RC
python3 - <<'PY'
import json, os, sys
raw = os.environ.get("RAW") or ""
gate = os.environ["GATE"]
check = os.environ["CHECK"]
target = os.environ["TARGET"]
findings = []
n = 0
for line in raw.splitlines():
    line = line.strip()
    if not line:
        continue
    try:
        obj = json.loads(line)
    except Exception:
        continue
    info = obj.get("info") or {}
    name = info.get("name") or obj.get("template-id") or obj.get("templateID") or "nuclei-match"
    sev = (info.get("severity") or "medium").lower()
    if sev not in ("critical", "high", "medium", "low", "info"):
        sev = "medium"
    findings.append({
        "code": "sec.nuclei.match",
        "gate": gate,
        "check": check,
        "severity": sev,
        "message": str(name)[:240],
        "target": obj.get("matched-at") or obj.get("host") or target,
        "evidence": {
            "template": str(obj.get("template-id") or obj.get("templateID") or ""),
        },
    })
    n += 1
    if n >= 25:
        break

findings.append({
    "code": "sec.nuclei.completed" if n else "sec.nuclei.ok",
    "gate": gate,
    "check": check,
    "severity": "info",
    "message": f"nuclei completed with {n} match(es)" if n else "nuclei: no wordpress-tag matches",
    "target": target,
    "evidence": {"matches": str(n), "rc": os.environ.get("RC", "")},
})
print(json.dumps({"findings": findings}))
PY
