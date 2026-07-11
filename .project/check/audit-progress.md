# Check coverage progress

**SSOT for product checklists** in [theme-check.md](./theme-check.md), [validate-check.md](./validate-check.md), [security-check.md](./security-check.md).  
Update this file when a check lands in code. Cycle roadmap stays in [../PROGRESS.md](../PROGRESS.md).

**Legend:** `[x]` implemented · `[~]` partial / scaffold · `[ ]` not started  
**Last audited:** 2026-07-11 (Quality Q3–Q6 + audit.json emit)

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
| Median of 3 runs | [ ] | single run today |
| Resource byte budgets | [ ] | |
| Core Web Vitals asserts (LCP/CLS/INP) | [ ] | inside LH report unused |
| Target = WP theme (not only static fixture) | [ ] | L0 uses `static` adapter |

### Q2 — W3C HTML5 (vnu)

| Check | Status | Notes |
|-------|--------|-------|
| vnu Docker runner | [x] | `runners/vnu` → `lab/vnu:local`; registry `vnu` |
| 0 errors on chrome/controlled URLs | [x] | findings `quality.vnu.*`; gate in `quality.lab.yaml` |
| Soft mode for Unit Test content | [ ] | |

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
| Latte chrome-specific asserts (sheet trap, etc.) | [x] | keyboard chrome via Org C4 `org-keyboard`; focus-trap still deferred |

### Q5 — SEO meta / graph

| Check | Status | Notes |
|-------|--------|-------|
| title / viewport asserts | [x] | `seo-meta` runner |
| meta description | [x] | soft info if missing |
| Open Graph / Twitter cards (optional profile) | [~] | gated by `seoSocial=true` |
| JSON-LD parse (optional) | [~] | count only when `seoSocial=true` |

### Q6 — Modern extras

| Check | Status | Notes |
|-------|--------|-------|
| Viewports 360/768/1280 | [x] | `runners/quality-extras` |
| Console clean | [x] | pageerror + console.error |
| `prefers-reduced-motion` | [ ] | deferred |
| Broken-link crawl | [ ] | deferred |

### Quality infra

| Check | Status | Notes |
|-------|--------|-------|
| Static fixture adapter | [x] | `adapters/static` + `quality-site` |
| Compose `quality` + nginx fixture | [x] | profile `quality` |
| Policy pack `lightspeed` | [x] | includes vnu/css/seo/extras |
| Manifest `quality.lab.yaml` (Q1–Q6) | [x] | |
| Logged-out prod-like WP target | [ ] | |

---

## Lab `sec` ← security-check.md

### S1 — Recon

| Check | Status | Notes |
|-------|--------|-------|
| `readme.html` exposed | [x] | `sec.recon.readme` |
| XML-RPC `system.listMethods` | [x] | `sec.recon.xmlrpc` |
| Generator / version leak | [ ] | |
| User enum (`?author=`, REST users) | [ ] | |
| Registration open | [ ] | |
| REST index surface | [ ] | |
| Directory listing uploads/themes | [ ] | |
| Sensitive files (`.env`, `wp-config.bak`, `debug.log`) | [ ] | |
| wp-cron abuse note | [ ] | |
| WPScan enum u,t,p (Docker) | [~] | wrapper emits `sec.wpscan.completed`, weak CVE map |

### S2 — Known CVE

| Check | Status | Notes |
|-------|--------|-------|
| WPScan API token CVE match | [ ] | |
| Nuclei WP templates | [ ] | |
| `composer audit` on theme | [ ] | |

### S3 — Auth abuse

| Check | Status | Notes |
|-------|--------|-------|
| Limited password spray | [ ] | lab-only |
| XML-RPC multicall flood detect | [ ] | |
| Cookie flags after login | [ ] | |

### S4 — Theme attack surface

| Check | Status | Notes |
|-------|--------|-------|
| Static dangerous PHP patterns (eval, unserialize, …) | [~] | partial via zip policy scan only |
| Semgrep / phpcs-security | [ ] | |
| Dynamic XSS fixtures | [ ] | |

### S5 — Headers / config

| Check | Status | Notes |
|-------|--------|-------|
| `X-Content-Type-Options` | [x] | `sec.headers.nosniff` |
| `X-Frame-Options` / CSP frame-ancestors | [x] | `sec.headers.clickjacking` |
| `Referrer-Policy` | [x] | `sec.headers.referrer` |
| CSP / HSTS / Permissions-Policy | [ ] | |
| `DISALLOW_FILE_EDIT` probe | [ ] | needs WP admin/config |

### Sec infra

| Check | Status | Notes |
|-------|--------|-------|
| Policy pack `secure-baseline` | [x] | |
| Manifest `sec.lab.yaml` | [x] | |
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
| Registry adapters/runners | [x] | includes `vnu` |
| Demo lab | [x] | |
| WordPress adapter stub | [x] | baseUrl + themeZip + matrix |
| wpfasty `theme:verify` / `theme:audit` client | [~] | `theme:audit` → Lab org `-o dist/*.audit.json`; full verify later |
| SaaS API / workers | [ ] | Cycle F |

---

## Suggested next (priority)

1. **Sec S1** — user enum + sensitive files + REST  
2. **Q1 deepen** — median-of-3 / CWV / WP target for quality  
3. **Q6 deferred** — prefers-reduced-motion + broken-link crawl  
4. Cycle E — static-web adapters  

---

## Counts (approximate)

| Area | Done | Partial | Todo |
|------|------|---------|------|
| Org Gate 1 | 26 | 0 | 0 |
| Org Gate 2–4 | 18 | 0 | 0 |
| Quality Q1–Q6 | 20 | 2 | ~6 |
| Sec S1–S5 | 6 | 2 | ~15 |
| Framework | 8 | 0 | 2 |
