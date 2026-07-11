# Cycle F — worker isolation & golden exits (F2.5 / F2.6)

## Isolation (per-run policy)

Until multi-tenant queues land, one API worker process runs jobs **serially** (`LAB_API_WORKERS=1` default).

| Concern | Policy |
|---------|--------|
| Docker networks | Org/theme-check join compose network `fastygo-lab_lab` + volume `fastygo-lab_wp_org_data`. Browser/LH/sec scanners use `dockerNetwork: host` + public/lab `baseUrl`. |
| Shared WP | Org / quality-wp / sec share the **same** compose WordPress. Do not run two mutating jobs (org theme-check + sec) in parallel on one host. |
| Theme zip | Bound via POST `themeZip` or manifest; mounted read-only into runners. Prefer `testdata/dist/*.zip` on the worker host. |
| Cleanup | Adapter `Teardown` stops static servers. Compose stack stays up between jobs (operator-owned). No automatic volume wipe. |
| Disk | Runner images + Chromium + WP volume — budget ≥40 GB on VPS. Job timeout: API worker uses 60 min context. |
| Secrets | `LAB_WP_PASSWORD`, `LAB_DATABASE_URL`, notify tokens via env only — never in committed manifests. |

## Owned-target allowlist (sec)

WordPress adapter honors `LAB_ALLOWED_BASE_URLS` (comma-separated URL prefixes). When set, `baseUrl` must match one prefix or Prepare fails.

Example on the lab VPS:

```bash
export LAB_ALLOWED_BASE_URLS='http://127.0.0.1:8080,http://5.129.242.217:8080'
```

Unset = no restriction (local CLI). SaaS workers **should** set this.

## Golden exits (known fail vs error)

| Outcome | Meaning | HTTP/API `status` |
|---------|---------|-------------------|
| Orchestrator crash / unknown runner | Infra bug | `error` |
| Checks ran; policy has high/critical | Product fail | `fail` |
| Checks ran; only warn baskets | Soft | `warn` |
| Checks ran; clean | Green | `pass` |
| Fixture budgets / axe target-size on static quality | **Known fail** for tight L0 fixture — still a successful job | `fail` + report |

CLI exit codes follow the same Report status. SaaS clients should treat `error` as retryable infra; `fail`/`warn`/`pass` as completed audits.
