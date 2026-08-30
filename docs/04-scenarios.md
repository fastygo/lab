# 04 — Product scenarios

Three scenarios = three product questions. Use them as a **curriculum**, not as random Docker scripts.

```text
theme:verify ≈ org + sec + quality
```

---

## Scenario A — Org (`org`)

**Question:** *Would a wordpress.org Theme Reviewer bounce this zip on packaging / Theme Check / template smoke / basic keyboard?*

| Gate | What runs | Why juniors care |
|------|-----------|------------------|
| **C1 Zip lint** | `zip-lint` | Catches missing `readme.txt`, bad screenshot, forbidden files — before Docker |
| **C2 Theme Check** | `theme-check` in clean WP | Closest automated stand-in for Theme Check plugin |
| **C3 Matrix** | `http-matrix` + `notice-hunter` | Front/single/404/… return expected status; no theme PHP notices |
| **C4 Keyboard** | `org-keyboard` (Playwright) | Skip link, nav, mobile sheet, search |

**Preset:** `org`  
**Manifest:** `testdata/manifests/org.lab.yaml`  
**Needs:** `make org-up`, theme zip, `lab/theme-check:local` (+ keyboard image)

### Upstream standards (read these)

- [Theme Handbook — Review](https://make.wordpress.org/themes/handbook/review/)  
- [Theme Handbook — Required](https://make.wordpress.org/themes/handbook/review/required/)  
- [Theme Check plugin](https://wordpress.org/plugins/theme-check/)  
- [Theme Unit Test data](https://codex.wordpress.org/Theme_Unit_Test) / [theme-test-data](https://github.com/WordPress/theme-test-data)

Design deep-dive: [.project/check/theme-check.md](../.project/check/theme-check.md).

### Typical decisions

| Finding class | Basket |
|---------------|--------|
| Missing license / bad text domain | Fix in theme zip |
| Theme Check **required** error | Fix in theme |
| `accessibility-ready` tag without focus trap | Block tag (do not claim yet) |
| Notice from Unit Test *content* | Often ACCEPT / soft — do not hard-fail the whole theme on user HTML |

---

## Scenario B — Quality (`quality` / `quality-wp`)

**Question:** *Is the logged-out front-end fast enough, valid enough, accessible enough, and SEO-sane?*

| Gate | Runner(s) | Notes |
|------|-----------|--------|
| **Q1 Lighthouse** | `lighthouse` | Mobile categories + budgets + CWV proxies (LCP/CLS/TBT) |
| **Q2 HTML** | `vnu` | [Nu Html Checker](https://validator.github.io/validator/) — soft mode on WP content |
| **Q3 CSS** | `css-lint` | Parse / stylelint hygiene on theme CSS |
| **Q4 axe** | `axe` | WCAG-oriented [axe-core](https://github.com/dequelabs/axe-core) rules |
| **Q5 SEO meta** | `seo-meta` | Title/viewport/h1; social optional |
| **Q6 extras** | `quality-extras` | Viewports, console smoke |

**Presets:**

| Preset | Target |
|--------|--------|
| `quality` | Static fixture (`testdata/fixtures/quality-site`) — fast iteration |
| `quality-wp` | Live WP logged-out (honest theme scores) |

**Needs:** runner images; for WP path — org WordPress up + zip.

### Upstream standards

- [web.dev / Lighthouse](https://developer.chrome.com/docs/lighthouse/overview/)  
- [Core Web Vitals](https://web.dev/articles/vitals)  
- [WCAG 2.2](https://www.w3.org/TR/WCAG22/)  
- [Nu Html Checker](https://validator.w3.org/nu/)  
- [axe Rules](https://github.com/dequelabs/axe-core/blob/develop/doc/rule-descriptions.md)

Design deep-dive: [.project/check/validate-check.md](../.project/check/validate-check.md).

### Typical decisions

| Finding | Basket |
|---------|--------|
| Huge theme CSS/JS | FIX_THEME (trim) or BUDGET (marketing preset) |
| Missing image dimensions / CLS | FIX_THEME |
| Hosting compression / HTTP/2 | SITE / FIX_SITE class |
| Unit Test malformed HTML | soft / BUDGET — not a hard theme fail |
| Claiming `accessibility-ready` | BLOCK_TAG until Q4 + keyboard + manual SR |

**Lab rule:** run quality **logged-out** (no admin bar) or Perf/HTML lie.

---

## Scenario C — Security (`sec`)

**Question:** *What attack surface do we accept, cut from the theme, or harden on the site baseline?*

| Gate | Runner(s) | Notes |
|------|-----------|--------|
| **S1 Recon** | `headers` (recon probes) | generator, xmlrpc, user enum, sensitive files, … |
| **S2 CVE** | `wpscan`, `composer-audit`, `nuclei` | Known vulns + deps |
| **S3 Auth abuse** | `auth-abuse` | Limited spray, xmlrpc multicall, cookie flags, host-header |
| **S4 Theme** | `theme-sec`, `phpcs-security`, `semgrep` | Dangerous PHP / Latte patterns |
| **S5 Headers** | covered with recon/headers | CSP, nosniff, frame, … |

**Preset:** `sec`  
**Needs:** owned WP URL, allowlist, scanner images. Optional `WPSCAN_API_TOKEN` for richer CVE data.

### Hard rule

**Only scan owned / lab targets.** On SaaS workers set:

```bash
LAB_ALLOWED_BASE_URLS=http://127.0.0.1:8080,http://YOUR_HOST:8080
```

Never point sec presets at customer production without written approval.

### Upstream standards

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)  
- [OWASP WordPress Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/WordPress_Security_Cheat_Sheet.html)  
- [WPScan](https://wpscan.com/) / [wpscanteam/wpscan](https://github.com/wpscanteam/wpscan)  
- [ProjectDiscovery Nuclei](https://docs.projectdiscovery.io/tools/nuclei/overview)  
- [PHPCS Security Audit](https://github.com/FloeDesignTechnologies/phpcs-security-audit)  
- [Semgrep](https://semgrep.dev/docs/)

Design deep-dive: [.project/check/security-check.md](../.project/check/security-check.md).

### Typical decisions

| Finding | Basket |
|---------|--------|
| XSS risk in theme PHP / `|noescape` | CUT_THEME |
| XML-RPC / user enum open | SITE_DEFAULT_OFF |
| Missing security headers | SITE_DEFAULT_ON |
| Login rate limit absent | SITE_DEFAULT_ON (mu-plugin — **not** in .org zip) |
| REST index visible | often ACCEPT (Gutenberg) |

**.org themes** must stay thin — login limiters / WAF belong on the **site blueprint**, not in the theme zip.

---

## Which scenario when?

| Situation | Run |
|-----------|-----|
| PR touching packaging / templates | `org` (or zip-lint alone via CLI manifest) |
| Perf / a11y regression on front | `quality-wp` |
| Static HTML playground | `quality` or `static-web` |
| Before release / weekly hardening review | `sec` |
| Full constructor “ready?” | all three (serial on one WP host) |

**Isolation tip:** do not run two **mutating** jobs (theme install + aggressive sec) in parallel on the same WordPress. Default API worker is serial (`LAB_API_WORKERS=1`) for that reason.

Next: [05 — API & dashboard](./05-api-dashboard.md).
