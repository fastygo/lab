#!/bin/sh
set -eu
# Thin wrapper around wpscanteam/wpscan — maps JSON output to findings.
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

# WPScan image is Ruby-based — parse with ruby.
export RAW GATE CHECK TARGET
ruby -e '
require "json"
raw = ENV["RAW"]
gate = ENV["GATE"]
check = ENV["CHECK"]
target = ENV["TARGET"]
findings = []
begin
  data = JSON.parse(raw)
rescue => e
  puts({findings:[{code:"sec.wpscan.exec_failed",gate:gate,check:check,severity:"high",message:"invalid wpscan JSON: #{e}",target:target}]}.to_json)
  exit 0
end

users = data["users"] || {}
if users.is_a?(Hash) && !users.empty?
  findings << {
    code: "sec.wpscan.users",
    gate: gate, check: check,
    severity: "medium",
    message: "WPScan enumerated #{users.size} user(s)",
    target: target,
    evidence: {count: users.size.to_s},
  }
end

vulns = []
version = data["version"] || {}
vulns.concat(version["vulnerabilities"] || []) if version.is_a?(Hash)
%w[themes plugins].each do |key|
  bag = data[key]
  next unless bag.is_a?(Hash)
  bag.each_value do |item|
    next unless item.is_a?(Hash)
    vulns.concat(item["vulnerabilities"] || [])
  end
end

vulns.first(20).each do |v|
  title = (v["title"] || v["id"] || "vulnerability").to_s[0, 240]
  findings << {
    code: "sec.wpscan.vuln",
    gate: gate, check: check,
    severity: "high",
    message: title,
    target: target,
  }
end

findings << {
  code: "sec.wpscan.completed",
  gate: gate, check: check,
  severity: "info",
  message: "wpscan enumeration completed",
  target: target,
  evidence: {
    bytes: raw.bytesize.to_s,
    vulns: vulns.size.to_s,
    users: (users.is_a?(Hash) ? users.size : 0).to_s,
  },
}
puts({findings: findings}.to_json)
'
