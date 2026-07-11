# Runners

A **runner** executes one tool (or a small tool family) and emits **findings** JSON.

## Built-in (Cycles B–D)

| Id | Type | Path |
|----|------|------|
| `stub` | in-process | `packages/orchestrator/stub` |
| `zip-lint` | in-process | `packages/orchestrator/ziplint` |
| `headers` | in-process | S1 recon + S5 headers |
| `auth-abuse` | in-process | S3 spray / xmlrpc multicall / cookies |
| `theme-sec` | in-process | S4 zip danger patterns + search XSS |
| `seo-meta` | in-process | `packages/orchestrator/seometa` |
| `http-matrix` | in-process | `packages/orchestrator/httpmatrix` |
| `lighthouse` | Docker | `runners/lighthouse` → `lab/lighthouse:local` |
| `axe` | Docker | `runners/axe` → `lab/axe:local` |
| `theme-check` | Docker | `runners/theme-check` → `lab/theme-check:local` |
| `vnu` | Docker | `runners/vnu` → `lab/vnu:local` |
| `wpscan` | Docker | `runners/wpscan` → `lab/wpscan:local` |
| `composer-audit` | Docker | `runners/composer-audit` → `lab/composer-audit:local` |
| `nuclei` | Docker | `runners/nuclei` → `lab/nuclei:local` |
| `phpcs-security` | Docker | `runners/phpcs-security` → `lab/phpcs-security:local` |
| `semgrep` | Docker | `runners/semgrep` → `lab/semgrep:local` |
| `notice-hunter` | Docker | `runners/notice-hunter` → `lab/notice-hunter:local` |
| `org-keyboard` | Docker | `runners/org-keyboard` → `lab/org-keyboard:local` |
| `css-lint` | Docker | `runners/css-lint` → `lab/css-lint:local` |
| `quality-extras` | Docker | `runners/quality-extras` → `lab/quality-extras:local` |

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
