# Lab: org — WordPress.org readiness

**Id:** `org`  
**Cycle:** C  
**Adapters:** `wordpress` (capabilities: `http`, `html`, `wp-theme`)

## Purpose

Prove a classic theme zip is ready for WordPress.org Theme Directory review.

## Gates (summary)

| Gate | Checks |
|------|--------|
| Zip lint | Required files, forbidden paths, screenshot 4:3 ≤1200×900, tags honesty, Resources coverage |
| Theme Check | 0 required errors on installed zip |
| Templates | URL matrix: front, home, single, page, category, tag, author, search, 404, attachment |
| Debug clean | WP_DEBUG log: no theme notices/warnings |
| Keyboard | Skip link, menus, mobile sheet, search (overlap with quality) |

## Out of scope

- WooCommerce, FSE, comments/PII forms, RTL (product exclusions unless lab config opts in)
- Claiming `accessibility-ready` without focus-trap + manual pass

## Runners

- `zip-lint` — Gate 1 packaging
- `theme-check` — Gate 2 headless Theme Check (install zip on compose WP)
- `http-matrix` — Gate 3 HTTP smoke GET asserts on adapter URL matrix
- optional Playwright keyboard (Gate 4)

## Policy pack

`wordpress-org` → `FIX_THEME`, `BLOCK_TAG`, `ACCEPT` as appropriate.

## Status

- **zip-lint** — full Gate 1 (see `.project/check/theme-check.md`)
- **theme-check** — Docker runner installs `themeZip`, activates Theme Check, emits `org.themecheck.*` (CLI or `run-check.php`)
- **http-matrix** — real HTTP status asserts (`org.matrix.ok` / `status_*`); `listOnly=true` for list-only mode
- Manifest: `testdata/manifests/org.lab.yaml` (`themeZip: testdata/dist/latte.zip`)
- Policy pack: `wordpress-org`

Config knobs on zip-lint: `allowAccessibilityReady`, `allowEcommerce`, `skipPolicyScan`.

Gate 2 check config: `dockerNetwork`, `wpDataVolume`, `internalUrl` (defaults in `org.lab.yaml` for compose project `fastygo-lab`).

```bash
make org-up && make runners
go run ./apps/cli run -f testdata/manifests/org.lab.yaml
```


