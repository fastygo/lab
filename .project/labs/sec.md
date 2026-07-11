# Lab: sec — security / hardening

**Id:** `sec`  
**Cycle:** D  
**Adapters:** `wordpress` or generic `http`

## Purpose

Attack **owned lab targets** to decide what to cut from artifacts vs enable/disable on site baseline.

## Gates (summary)

| Gate | Focus |
|------|-------|
| S1 Recon | Version leak, user enum, xmlrpc, REST, sensitive files, directory listing |
| S2 CVE | WPScan (+ API token), `composer audit` on theme zip, Nuclei wordpress tags |
| S3 Auth abuse | Limited spray, xmlrpc multicall, cookie flags (`LAB_WP_PASSWORD`), host-header reset |
| S4 Target SAST/dynamic | Theme zip danger patterns + reflected search XSS probe |
| S5 Headers/config | Security headers, `DISALLOW_FILE_EDIT` endpoint note (via `headers` runner) |

## Decision baskets

`CUT_TARGET`, `SITE_DEFAULT_ON`, `SITE_DEFAULT_OFF`, `FIX_THEME`, `FIX_SITE`, `ACCEPT`

## Hard rules

- No scanning third-party production sites without explicit opt-in
- Do not ship login-limiter/WAF inside .org theme zips — site baseline only
- Runners emit findings; policy assigns baskets
- Password spray uses **fake** passwords only (`lab-spray-never-*`) — measures lockout, does not crack

## Status (Cycle D)

- **headers** — S1 recon + S5 header scorecard
- **wpscan** / **composer-audit** / **nuclei** — S2
- **auth-abuse** — S3
- **theme-sec** — S4
- Manifest: `testdata/manifests/sec.lab.yaml`
- Policy pack: `secure-baseline`

Open: Semgrep/PHPCS-security, deeper XSS fixtures, SameSite raw cookie assert.

