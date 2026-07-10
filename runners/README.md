# Runners

A **runner** executes one tool (or a small tool family) and emits **findings** JSON.

## Built-in (Cycles B–D)

| Id | Type | Path |
|----|------|------|
| `stub` | in-process | `packages/orchestrator/stub` |
| `zip-lint` | in-process | `packages/orchestrator/ziplint` |
| `headers` | in-process | `packages/orchestrator/headers` |
| `http-matrix` | in-process | `packages/orchestrator/httpmatrix` |
| `lighthouse` | Docker | `runners/lighthouse` → `lab/lighthouse:local` |
| `axe` | Docker | `runners/axe` → `lab/axe:local` |
| `theme-check` | Docker | `runners/theme-check` → `lab/theme-check:local` |
| `wpscan` | Docker | `runners/wpscan` / `wpscanteam/wpscan` |

Build: `make runners`

If Docker is missing, Docker-backed runners emit finding `runner.docker.unavailable` (severity high) instead of panicking.

## Contract

### Input

| Key | Meaning |
|-----|---------|
| `LAB_TARGET_URL` | Base URL of the target |
| `LAB_CHECK_ID` | Check id from the manifest |
| `LAB_GATE_ID` | Gate id |
| `LAB_CONFIG_JSON` | Check-specific config |
| `LAB_THEME_ZIP` | Optional theme zip path |
| `WPSCAN_API_TOKEN` | Optional WPScan API token |

### Output

Stdout JSON: `{ "findings": [ ... ] }` or a raw findings array.  
Exit `0` = runner completed (findings may be non-empty).
