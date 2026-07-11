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
| `vnu` | Docker | `runners/vnu` → `lab/vnu:local` |
| `wpscan` | Docker | `runners/wpscan` / `wpscanteam/wpscan` |
| `notice-hunter` | Docker | `runners/notice-hunter` → `lab/notice-hunter:local` |
| `org-keyboard` | Docker | `runners/org-keyboard` → `lab/org-keyboard:local` |
| `css-lint` | Docker | `runners/css-lint` → `lab/css-lint:local` |
| `quality-extras` | Docker | `runners/quality-extras` → `lab/quality-extras:local` |
| `seo-meta` | in-process | `packages/orchestrator/seometa` |

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
| `LAB_THEME_ZIP` | Optional theme zip path (container path after mount) |
| `WPSCAN_API_TOKEN` | Optional WPScan API token |

Theme Check (Gate 2) also uses check config: `dockerNetwork`, `wpDataVolume`, `internalUrl`. The Go docker runner mounts host `themeZip` to `/lab/theme.zip`. CSS lint uses `cssDir` → `/lab/css`.

### Output

Stdout JSON: `{ "findings": [ ... ] }` or a raw findings array.  
Exit `0` = runner completed (findings may be non-empty).

CLI also supports writing the full lab report:

```bash
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml -o /tmp/quality.audit.json
```
