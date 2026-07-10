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

## Runners (planned)

- `zip-lint`
- `theme-check`
- `http-matrix` / wp-cli smoke
- optional Playwright keyboard

## Policy pack

`wordpress-org` → `FIX_THEME`, `BLOCK_TAG`, `ACCEPT` as appropriate.

## Status (Cycle C + Gate 1 P0)

- **zip-lint** implements full Gate 1 from `.project/check/theme-check.md`:
  - required files (`style.css`, `readme.txt`, `LICENSE`, `screenshot`, `functions.php`)
  - forbidden paths (`.git`, `.cursor`, `node_modules`, monorepo dirs, OS junk)
  - forbidden extensions (`.xml` allowlist, `.sh`, `.sql`, nested `.zip`)
  - screenshot ≤1200×900 ~4:3
  - `style.css` headers: Version, Requires at least, Tested up to, Requires PHP, Text Domain **= folder slug**
  - tags: block `accessibility-ready` / `e-commerce` unless config allows
  - Resources section vs `assets/` + `lib/` attribution
  - minified twin rule (`*.min.js/css` → source)
  - policy scan: CPT, shortcode, woocommerce, comments_template, contact-form patterns
- **theme-check** Docker/compose path (`runners/theme-check`, profile `org`)
- **http-matrix** records adapter URL matrix
- Manifest: `testdata/manifests/org.lab.yaml`
- Policy pack: `wordpress-org`

Config knobs on the check: `allowAccessibilityReady`, `allowEcommerce`, `skipPolicyScan`.

Set `spec.adapter.config.themeZip` to a theme zip path before expecting zip-lint to pass.


