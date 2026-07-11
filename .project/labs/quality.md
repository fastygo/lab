# Lab: quality — performance, standards, a11y, SEO

**Id:** `quality`  
**Cycle:** B  
**Adapters:** `static` (fixture) + `wordpress` (logged-out runtime)

## Purpose

Validate modern site quality: Lighthouse (4×100), W3C HTML5, CSS hygiene, ARIA/WCAG, optional SEO graph.

## Gates (summary)

| Gate | Tooling | Default thresholds (theme) |
|------|---------|----------------------------|
| Q1 Lighthouse | median×3 + CWV + byte budgets | Perf / A11y / BP / SEO + LCP/CLS/TBT + bytes |
| Q2 HTML5 | vnu | hard on fixture; `softMode` on WP Unit Test content |
| Q3 CSS | stylelint / parse | 0 fatal parse errors |
| Q4 ARIA/WCAG | axe WCAG 2.2 AA | 0 critical/serious |
| Q5 SEO graph | seo-meta | title/viewport/h1; OG/Twitter/JSON-LD if `seoSocial=true` |
| Q6 Modern extras | Playwright | viewports, console, reduced-motion, links |

## Lab rules

- Logged-out, prod-like (`WP_DEBUG` display off)
- Do not hard-fail on arbitrary Unit Test content HTML; prefer controlled chrome URLs for vnu hard gate; WP uses `softMode`
- `accessibility-ready` tag remains `BLOCK_TAG` until focus-trap + manual SR (Org, not Quality)

## Manifests

| File | Adapter | Notes |
|------|---------|-------|
| `testdata/manifests/quality.lab.yaml` | `static` | fixture L0; `seoSocial=true`; tight byte budgets |
| `testdata/manifests/quality-wp.lab.yaml` | `wordpress` | `:8080` latte; soft vnu; looser budgets |

```bash
make quality-up   # nginx fixture :8091
make org-up       # WP :8080 for quality-wp
make runners
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml -o /tmp/quality.audit.json
go run ./apps/cli run -f testdata/manifests/quality-wp.lab.yaml -o /tmp/quality-wp.audit.json
```

Honestly out of Quality scope: field INP (CrUX), sheet focus-trap / `accessibility-ready`.
