# Check coverage progress

**SSOT for product checklists** in [theme-check.md](./theme-check.md), [validate-check.md](./validate-check.md), [security-check.md](./security-check.md).  
Update this file when a check lands in code. Cycle roadmap stays in [../PROGRESS.md](../PROGRESS.md).

**Legend:** `[x]` implemented · `[~]` partial / scaffold · `[ ]` not started  
**Last audited:** 2026-07-11 (Quality checklist closed: budgets + WP + soft vnu + social)

---

## How to run what exists

```bash
cd d:/FastyGo/Lab
go test ./...
go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml   # Docker for real LH/axe/vnu
# org: themeZip=testdata/dist/latte.zip; Gate 2 needs compose org + theme-check image
make org-up && make runners
go run ./apps/cli run -f testdata/manifests/org.lab.yaml
go run ./apps/cli run -f testdata/manifests/sec.lab.yaml
```

---

## Lab `org` ← theme-check.md

### Gate 1 — Zip / packaging (`zip-lint`)

| Check | Status | Code / notes |
|-------|--------|----------------|
| Forbidden paths (`.git`, `.cursor`, `node_modules`, monorepo dirs, OS junk) | [x] | `org.zip.forbidden_path` |
| Required: `style.css`, `readme.txt`, `functions.php` | [x] | `org.zip.missing_*` |
| Required: `LICENSE` / `license.txt` | [x] | `org.zip.missing_license` |
| Required: `screenshot.png` | [x] | `org.zip.missing_screenshot` |
| Forbidden ext `.xml` (allowlist) | [x] | `org.zip.forbidden_ext_xml` |
| Forbidden ext `.sh`, `.sql` | [x] | `org.zip.forbidden_ext_sh/sql` |
| Nested zip-in-zip | [x] | `org.zip.nested_zip` |
| Screenshot ≤1200×900 | [x] | `org.zip.screenshot_size` |
| Screenshot ratio ~4:3 | [x] | `org.zip.screenshot_ratio` |
| `style.css` Version | [x] | `org.zip.style_version` |
| `style.css` Requires at least | [x] | `org.zip.style_requires` |
| `style.css` Tested up to | [x] | `org.zip.style_tested` |
| `style.css` Requires PHP | [x] | `org.zip.style_requires_php` |
| `style.css` Text Domain | [x] | `org.zip.style_textdomain` |
| Text Domain = folder slug | [x] | `org.zip.style_textdomain_slug` |
| Block tag `accessibility-ready` | [x] | `org.zip.tag_accessibility_ready` (+ config override) |
| Block tag e-commerce / Woo | [x] | `org.zip.tag_ecommerce` |
| Resources section if `assets/`/`lib/` | [x] | `org.zip.resources_missing_section` |
| Resources attribution per asset | [x] | `org.zip.resources_unattributed` |
| Minified twin (`*.min.js/css`) | [x] | `org.zip.minified_without_source` |
| Policy: `register_post_type` | [x] | `org.zip.policy_cpt` |
| Policy: `add_shortcode` | [x] | `org.zip.policy_shortcode` |
| Policy: woocommerce refs | [x] | `org.zip.policy_woocommerce` |
| Policy: `comments_template` | [x] | `org.zip.policy_comments_template` |
| Policy: contact/newsletter patterns | [x] | `org.zip.policy_contact_form` |
| Emit `dist/*.audit.json` from wpfasty | [x] | `lab run -o`; wpfasty `bun run theme:audit -- <theme>` |

**Gate 1:** done in Lab (P0).

### Gate 2 — Clean WP + Theme Check

| Check | Status | Notes |
|-------|--------|-------|
| Compose profile `org` (db + wordpress + wpcli) | [x] | `deploy/compose`; `make org-up` |
| Install theme from zip (not monorepo bind) | [x] | docker runner mounts `themeZip` → `/lab/theme.zip`; unzip+activate |
| Install + activate Theme Check plugin | [x] | `runners/theme-check/entrypoint.sh` |
| Headless Theme Check → 0 required errors | [x] | CLI or `run-check.php` eval-file; findings `org.themecheck.*` |
| Parse required vs recommended | [x] | PHP `to-findings.php` + Go `themecheck.ParseJSON` |
| WP_DEBUG log capture for theme | [x] | compose `WP_DEBUG_LOG`; Gate 3c `notice-hunter` |

### Gate 3 — Theme Unit Test + templates

| Check | Status | Notes |
|-------|--------|-------|
| Import Theme Unit Test XML | [x] | `runners/theme-check/seed-unit-test.sh` + `make org-seed` |
| Cache XML in lab fixtures (not theme zip) | [x] | `testdata/fixtures/themeunittestdata.wordpress.xml` |
| HTTP smoke: front, home, single, page | [x] | `http-matrix` GET + status asserts |
| HTTP smoke: category, tag, author, search, 404, attachment | [x] | matrix + `org-seed.json` attachmentId |
| URL matrix listed on adapter | [x] | wordpress adapter `Matrix()` (reloaded per gate) |
| `http-matrix` records URLs as findings | [x] | real HTTP asserts (`org.matrix.ok` / `status_*`); `listOnly=true` for list mode |
| Notice hunter (`debug.log`) | [x] | runner `notice-hunter` → `org.notice.*` |

### Gate 4 — Keyboard / a11y chrome

| Check | Status | Notes |
|-------|--------|-------|
| Skip link first focusable + target `#content` | [x] | `runners/org-keyboard` → `org.keyboard.skip_*` |
| Primary nav keyboard | [x] | desktop viewport Tab into `header nav` |
| Mobile sheet open/close + aria | [x] | Escape closes; `aria-expanded` |
| Search focusable | [x] | Tab to `input[name=s]` |
| Block `accessibility-ready` until focus-trap | [x] | via zip-lint tag rule |

### Org policy / packaging

| Check | Status | Notes |
|-------|--------|-------|
| Policy pack `wordpress-org` | [x] | Gate 2/3 heuristics included |
| Manifest `org.lab.yaml` | [x] | `themeZip: testdata/dist/latte.zip` |

---

## Lab `quality` ← validate-check.md

### Q1 — Lighthouse (4 × 100)

| Check | Status | Notes |
|-------|--------|-------|
| Docker runner image + entrypoint | [x] | `runners/lighthouse` |
| Mobile categories Perf / A11y / BP / SEO | [x] | findings `quality.lighthouse.*` |
| Thresholds fail/warn from config | [x] | |
| Median of 3 runs | [x] | config `runs` (default 3) |
| Resource byte budgets | [x] | total/script/style transfer medians |
| Core Web Vitals asserts (LCP/CLS/INP) | [x] | LCP/CLS/TBT/FCP; TBT = INP proxy (field INP out of scope) |
| Target = WP theme (not only static fixture) | [x] | `quality-wp.lab.yaml` → wordpress `:8080` |

### Q2 — W3C HTML5 (vnu)

| Check | Status | Notes |
|-------|--------|-------|
| vnu Docker runner | [x] | `runners/vnu` → `lab/vnu:local`; registry `vnu` |
| 0 errors on chrome/controlled URLs | [x] | findings `quality.vnu.*`; gate in `quality.lab.yaml` |
| Soft mode for Unit Test content | [x] | `softMode=true` → `quality.vnu.soft_error` / BUDGET |

### Q3 — CSS

| Check | Status | Notes |
|-------|--------|-------|
| stylelint / CSS parse runner | [x] | `runners/css-lint` → `quality.css.*` |
| Fatal parse = fail | [x] | + forbidden `expression()` / `behavior:` |

### Q4 — ARIA / WCAG (axe)

| Check | Status | Notes |
|-------|--------|-------|
| Docker axe + Playwright | [x] | `runners/axe` (uses `browser.newContext`) |
| WCAG 2.2 AA tags | [x] | |
| Fail critical/serious | [x] | `quality.axe.*` |
| Latte chrome-specific asserts (sheet trap, etc.) | [x] | keyboard via Org C4; focus-trap stays Org/tag deferred |

### Q5 — SEO meta / graph

| Check | Status | Notes |
|-------|--------|-------|
| title / viewport asserts | [x] | `seo-meta` runner |
| meta description | [x] | soft info if missing |
| Open Graph / Twitter cards (optional profile) | [x] | `seoSocial=true` on static; WP default off |
| JSON-LD parse (optional) | [x] | parse + validate when `seoSocial=true` |

### Q6 — Modern extras

| Check | Status | Notes |
|-------|--------|-------|
| Viewports 360/768/1280 | [x] | `runners/quality-extras` |
| Console clean | [x] | pageerror + console.error |
| `prefers-reduced-motion` | [x] | emulate reduce; fail infinite anim |
| Broken-link crawl | [x] | same-origin `a[href]` via Playwright request |

### Quality infra

| Check | Status | Notes |
|-------|--------|-------|
| Static fixture adapter | [x] | `adapters/static` + `quality-site` |
| `static-web` adapter (Vite/SPA) | [x] | Cycle E; `quality-staticweb.lab.yaml` |
| Compose `quality` + nginx fixture | [x] | profile `quality` |
| Policy pack `lightspeed` | [x] | bytes / soft vnu / social |
| Manifest `quality.lab.yaml` (Q1–Q6) | [x] | |
| Logged-out prod-like WP target | [x] | `quality-wp.lab.yaml` |

---

## Lab `sec` ← security-check.md

### S1 — Recon

| Check | Status | Notes |
|-------|--------|-------|
| `readme.html` exposed | [x] | `sec.recon.readme` |
| XML-RPC `system.listMethods` | [x] | `sec.recon.xmlrpc` |
| Generator / version leak | [x] | `sec.recon.generator` |
| User enum (`?author=`, REST users) | [x] | `sec.recon.user_enum.author` / `.rest` |
| Registration open | [x] | `sec.recon.registration` |
| REST index surface | [x] | `sec.recon.rest_index` (ACCEPT — Gutenberg) |
| Directory listing uploads/themes | [x] | `sec.recon.dir_listing.*` |
| Sensitive files (`.env`, `wp-config.bak`, `debug.log`) | [x] | `sec.recon.sensitive.*` |
| wp-cron abuse note | [x] | `sec.recon.wp_cron` |
| WPScan enum u,t,p (Docker) | [x] | `sec.wpscan.users` / `.vuln` / `.completed` |

### S2 — Known CVE

| Check | Status | Notes |
|-------|--------|-------|
| WPScan API token CVE match | [x] | `vp,vt` when `WPSCAN_API_TOKEN` set; maps `sec.wpscan.vuln` |
| Nuclei WP templates | [x] | runner `nuclei` → `sec.nuclei.match` / `.ok` |
| `composer audit` on theme | [x] | runner `composer-audit` on theme zip lockfile |

### S3 — Auth abuse

| Check | Status | Notes |
|-------|--------|-------|
| Limited password spray | [x] | `sec.auth.login_no_rate_limit` / `.rate_limit_present` (fake wordlist only) |
| XML-RPC multicall flood detect | [x] | `sec.auth.xmlrpc_multicall` |
| Cookie flags after login | [x] | HttpOnly / Secure / **SameSite** via raw `Set-Cookie`; needs `LAB_WP_PASSWORD` |
| Host-header reset poison | [x] | `sec.auth.host_header_poison` / `.ok` |

### S4 — Theme attack surface

| Check | Status | Notes |
|-------|--------|-------|
| Static dangerous PHP patterns (eval, unserialize, …) | [x] | runner `theme-sec` → `sec.theme.*` (+ `|noescape` on `.latte`) |
| Semgrep / phpcs-security | [x] | runners `semgrep` + `phpcs-security` |
| Dynamic XSS fixtures | [x] | search script + attribute breakout probes |

### S5 — Headers / config

| Check | Status | Notes |
|-------|--------|-------|
| `X-Content-Type-Options` | [x] | `sec.headers.nosniff` |
| `X-Frame-Options` / CSP frame-ancestors | [x] | `sec.headers.clickjacking` |
| `Referrer-Policy` | [x] | `sec.headers.referrer` |
| CSP / HSTS / Permissions-Policy | [x] | `sec.headers.csp` / `.hsts` / `.permissions` |
| `DISALLOW_FILE_EDIT` probe | [x] | `sec.config.file_edit` (endpoint present → SITE_DEFAULT_ON) |

### Sec infra

| Check | Status | Notes |
|-------|--------|-------|
| Policy pack `secure-baseline` | [x] | |
| Manifest `sec.lab.yaml` | [x] | S1–S5: headers, S2 wpscan+composer+nuclei, S3 auth, S4 theme-sec |
| Compose profile `sec` | [x] | shares wordpress service |
| Decision baskets documented | [x] | `.project/policy.md` |

---

## Framework (all labs)

| Item | Status | Notes |
|------|--------|-------|
| Domain Manifest / Finding / Report | [x] | |
| Policy engine + packs | [x] | default, lightspeed, wordpress-org, secure-baseline |
| Orchestrator + CLI `lab` | [x] | v0.2.0 |
| Docker runner port + unavailable finding | [x] | zip mount + compose network/volume for theme-check |
| Registry adapters/runners | [x] | includes `vnu`, `composer-audit`, `nuclei` |
| Demo lab | [x] | |
| WordPress adapter stub | [x] | baseUrl + themeZip + matrix |
| `static-web` adapter (Cycle E) | [x] | dist SPA serve + optional vite preview |
| wpfasty `theme:verify` / `theme:audit` client | [~] | `theme:audit` → Lab org `-o dist/*.audit.json`; full verify later |
| SaaS API / workers | [~] | Cycle F — F3 SSE + F4 notify; schedules open; [cycle-f-saas.md](../vps/cycle-f-saas.md) |

---

## Suggested next (priority)

1. Cycle F — **F4.1–F4.2** schedules (cron → enqueue)  
2. Optional: deeper XSS Unit Test fixtures  

---

## Counts (approximate)

| Area | Done | Partial | Todo |
|------|------|---------|------|
| Org Gate 1 | 26 | 0 | 0 |
| Org Gate 2–4 | 18 | 0 | 0 |
| Quality Q1–Q6 | 28 | 0 | 0 |
| Sec S1–S5 | 28 | 0 | 0 |
| Framework | 9 | 0 | 1 |
