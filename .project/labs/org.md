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

## Related

Consumer: `wpfasty` `theme:pack` → feed zip into this lab.
