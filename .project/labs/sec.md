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
| S2 CVE | WPScan/nuclei + composer audit for theme deps |
| S3 Auth abuse | Login/xmlrpc rate, cookie flags (lab wordlists only) |
| S4 Target SAST/dynamic | Theme XSS surfaces, dangerous PHP patterns |
| S5 Headers/config | Security headers, `DISALLOW_FILE_EDIT`, uploads PHP exec |

## Decision baskets

`CUT_TARGET`, `SITE_DEFAULT_ON`, `SITE_DEFAULT_OFF`, `FIX_THEME`, `FIX_SITE`, `ACCEPT`

## Hard rules

- No scanning third-party production sites without explicit opt-in
- Do not ship login-limiter/WAF inside .org theme zips — site baseline only
- Runners emit findings; policy assigns baskets

## Status (Cycle D)

- **headers** in-process recon/headers runner
- **wpscan** Docker runner wrapper
- Manifest: `testdata/manifests/sec.lab.yaml`
- Policy pack: `secure-baseline` (xmlrpc/readme → `SITE_DEFAULT_OFF`, missing headers → `SITE_DEFAULT_ON`)

Nuclei full template pack deferred.

