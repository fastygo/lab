# composer-audit

Runs `composer audit --format=json` against `composer.lock` inside the theme zip.

## Image

```bash
docker build -t lab/composer-audit:local runners/composer-audit
```

## Env

| Variable | Meaning |
|----------|---------|
| `LAB_THEME_ZIP` | Mounted theme zip (`/lab/theme.zip`) |
| `LAB_GATE_ID` / `LAB_CHECK_ID` | Finding metadata |

## Findings

- `sec.composer.advisory` — known advisory (severity from Composer)
- `sec.composer.abandoned` — abandoned packages
- `sec.composer.ok` / `sec.composer.completed`
- `sec.composer.lock_missing` / `sec.composer.zip_missing` / `sec.composer.exec_failed`
