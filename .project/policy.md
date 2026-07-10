# Policy

Policy maps **Findings** → **Decisions** (baskets). Product owners configure packs; code stays generic.

## Baskets

| Basket | Meaning | Typical owner |
|--------|---------|---------------|
| `CUT_TARGET` | Remove feature/surface from the artifact under test | theme / app authors |
| `FIX_THEME` | Fix in WordPress theme / presentation layer | `wpfasty` / theme |
| `FIX_SITE` | Fix in site baseline / hosting / mu-plugin | site stack |
| `SITE_DEFAULT_ON` | Enable hardening by default on root site | secure baseline |
| `SITE_DEFAULT_OFF` | Disable risky surface by default | secure baseline |
| `BUDGET` | Adjust numeric budgets / accept with warn | lab preset |
| `ACCEPT` | Documented accepted risk | report only |
| `BLOCK_TAG` | Do not claim a marketplace tag (e.g. accessibility-ready) | packaging |

## Rules

- Policy does not run tools
- Same finding code can map differently per **policy pack** (`org`, `secure-baseline`, `lightspeed-blog`)
- Unmapped findings default to `ACCEPT` with severity preserved on the report
- Exit code: any `critical`/`high` without `ACCEPT` → fail (engine configurable)

## Packs (planned)

| Pack | Use |
|------|-----|
| `default` | Demo + generic |
| `wordpress-org` | Theme directory submission |
| `secure-baseline` | Production site hardening |
| `lightspeed` | Strict performance budgets |
