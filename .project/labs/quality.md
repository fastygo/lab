# Lab: quality — performance, standards, a11y, SEO

**Id:** `quality`  
**Cycle:** B  
**Adapters:** any with `http` + `html` (`wordpress`, `static-web`, …)

## Purpose

Validate modern site quality: Lighthouse (4×100), W3C HTML5, CSS hygiene, ARIA/WCAG, optional SEO graph.

## Gates (summary)

| Gate | Tooling | Default thresholds (theme) |
|------|---------|----------------------------|
| Q1 Lighthouse | LHCI mobile median | Perf ≥90 (fail &lt;70), A11y ≥95, BP ≥95, SEO ≥90 |
| Q2 HTML5 | vnu | 0 errors on chrome/controlled URLs |
| Q3 CSS | stylelint / parse | 0 fatal parse errors |
| Q4 ARIA/WCAG | axe WCAG 2.2 AA | 0 critical/serious |
| Q5 SEO graph | seo-meta runner | title/viewport/h1; OG if `seoSocial=true` |
| Q6 Modern extras | Playwright viewports + console | 360/768/1280 + console clean |

## Lab rules

- Logged-out, prod-like (`WP_DEBUG` display off)
- Do not hard-fail on arbitrary Unit Test content HTML; prefer controlled chrome URLs for vnu hard gate
- `accessibility-ready` tag remains `BLOCK_TAG` until focus-trap + manual SR

## L0 scope (implemented)

Gates: **Q1–Q6** via Docker runners + in-process `seo-meta` + `static` adapter fixture.

Deferred: broken-link crawl, `prefers-reduced-motion`, Q1 median-of-3 / CWV.

Manifest: `testdata/manifests/quality.lab.yaml`  
Policy pack: `lightspeed`

```bash
make quality-up
make runners   # includes css-lint + quality-extras
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml -o /tmp/quality.audit.json
```
