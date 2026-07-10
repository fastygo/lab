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
| Q5 SEO graph | optional | title/viewport; OG if profile enabled |
| Q6 Modern extras | Playwright viewports, console clean | pass |

## Lab rules

- Logged-out, prod-like (`WP_DEBUG` display off)
- Do not hard-fail on arbitrary Unit Test content HTML; prefer controlled chrome URLs for vnu hard gate
- `accessibility-ready` tag remains `BLOCK_TAG` until focus-trap + manual SR

## L0 scope (Cycle B — implemented)

Gates wired today: **Q1 Lighthouse**, **Q2 vnu**, **Q4 axe** via Docker runners + `static` adapter fixture.

Not yet: Q3 stylelint, Q5 SEO graph, Q6 extras.

Manifest: `testdata/manifests/quality.lab.yaml`  
Policy pack: `lightspeed`

