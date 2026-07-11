#!/bin/sh
set -eu
# Nu Html Checker (vnu) runner — Quality Q2.
GATE="${LAB_GATE_ID:-}"
CHECK="${LAB_CHECK_ID:-}"
TARGET="${LAB_TARGET_URL:-http://127.0.0.1/}"
SOFT_MODE=false
if [ -n "${LAB_CONFIG_JSON:-}" ]; then
  SOFT_MODE="$(LAB_CONFIG_JSON="$LAB_CONFIG_JSON" python3 -c 'import json,os; c=json.loads(os.environ.get("LAB_CONFIG_JSON") or "{}"); print("true" if str(c.get("softMode","")).lower() in ("1","true","yes") else "false")' 2>/dev/null || echo false)"
fi

TMP="$(mktemp)"
JSON_IN="$(mktemp)"
trap 'rm -f "$TMP" "$JSON_IN"' EXIT

if ! curl -fsSL --max-time 30 "$TARGET" -o "$TMP"; then
  printf '%s\n' "{\"findings\":[{\"code\":\"quality.vnu.fetch_failed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"failed to fetch $TARGET\",\"target\":\"$TARGET\"}]}"
  exit 0
fi

# Official image exposes `vnu` wrapper; fall back to jar path.
# Note: vnu emits JSON on stderr (stdout is empty).
OUT=""
if command -v vnu >/dev/null 2>&1; then
  OUT="$(vnu --format json --exit-zero-always "$TMP" 2>&1 || true)"
elif [ -f /vnu.jar ]; then
  OUT="$(java -jar /vnu.jar --format json --exit-zero-always "$TMP" 2>&1 || true)"
elif [ -f /opt/vnu/vnu.jar ]; then
  OUT="$(java -jar /opt/vnu/vnu.jar --format json --exit-zero-always "$TMP" 2>&1 || true)"
fi

# Keep only the JSON object if mixed with other text.
case "$OUT" in
  \{*) ;;
  *)
    OUT="$(printf '%s' "$OUT" | tr -d '\r' | sed -n '/^{/,$p')"
    ;;
esac

if [ -z "$OUT" ]; then
  printf '%s\n' "{\"findings\":[{\"code\":\"quality.vnu.exec_failed\",\"gate\":\"$GATE\",\"check\":\"$CHECK\",\"severity\":\"high\",\"message\":\"vnu produced no output\",\"target\":\"$TARGET\"}]}"
  exit 0
fi

printf '%s' "$OUT" > "$JSON_IN"
export LAB_GATE_ID="$GATE" LAB_CHECK_ID="$CHECK" LAB_TARGET_URL="$TARGET" LAB_VNU_JSON="$JSON_IN" LAB_VNU_SOFT="$SOFT_MODE"
python3 - <<'PY'
import json, os
gate = os.environ.get("LAB_GATE_ID", "")
check = os.environ.get("LAB_CHECK_ID", "")
target = os.environ.get("LAB_TARGET_URL", "")
path = os.environ.get("LAB_VNU_JSON", "")
soft = os.environ.get("LAB_VNU_SOFT", "false").lower() in ("1", "true", "yes")
try:
    with open(path, "r", encoding="utf-8") as f:
        data = json.load(f)
except Exception as e:
    print(json.dumps({"findings":[{"code":"quality.vnu.parse_failed","gate":gate,"check":check,"severity":"high","message":str(e),"target":target}]}))
    raise SystemExit(0)
msgs = data.get("messages") or []
findings = []
errors = 0
soft_errors = 0
for m in msgs:
    typ = (m.get("type") or "").lower()
    sev = "info"
    code = "quality.vnu.info"
    if typ == "error":
        if soft:
            sev = "medium"
            code = "quality.vnu.soft_error"
            soft_errors += 1
        else:
            sev = "high"
            code = "quality.vnu.error"
            errors += 1
    elif typ == "info" and (m.get("subType") or "") == "warning":
        sev = "medium"
        code = "quality.vnu.warning"
    msg = m.get("message") or ""
    findings.append({
        "code": code,
        "gate": gate,
        "check": check,
        "severity": sev,
        "message": msg,
        "target": target,
        "evidence": {
            "type": typ,
            "lastLine": str(m.get("lastLine") or ""),
            "softMode": "true" if soft else "false",
        },
    })
if not findings:
    findings.append({
        "code": "quality.vnu.ok",
        "gate": gate,
        "check": check,
        "severity": "info",
        "message": "vnu: 0 messages",
        "target": target,
    })
elif errors == 0:
    note = f"vnu: 0 hard errors"
    if soft and soft_errors:
        note = f"vnu softMode: {soft_errors} content error(s) demoted"
    elif soft:
        note = f"vnu softMode: 0 errors, {len(msgs)} messages"
    else:
        note = f"vnu: 0 errors, {len(msgs)} messages"
    findings.insert(0, {
        "code": "quality.vnu.no_errors",
        "gate": gate,
        "check": check,
        "severity": "info",
        "message": note,
        "target": target,
        "evidence": {"softMode": "true" if soft else "false", "softErrors": str(soft_errors)},
    })
print(json.dumps({"findings": findings}))
PY
