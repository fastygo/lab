# Compose profiles

Lab intensity levels for local Docker.

| Profile | Level | Intent | Typical contents |
|---------|-------|--------|------------------|
| `smoke` | L0 | PR-fast | noop/demo, maybe one cheap runner |
| `org` | L1/L2 | WP.org gate | Theme Check, zip lint, WP + theme zip |
| `sec` | L2 | Security lab | WPScan, nuclei, target WP |
| `quality` | L1/L2 | Perf/a11y/HTML | Lighthouse, axe, vnu |
| `full` | L2 | Release | org + sec + quality |
| `saas` | L3 | Cloud worker shape | same images, externalized config |

## Rules

- Profiles must not require the `wpfasty` monorepo bind-mount for release-shaped runs (install from zip/artifact).
- Dev bind-mounts are allowed only under an explicit `dev` profile (future).
- Network: dedicated bridge per compose project; runners resolve target by service DNS name.
- Cycle A: compose file documents profiles; real tool services arrive in B+.

## Commands (future)

```bash
docker compose -f deploy/compose/docker-compose.yml --profile smoke up -d
docker compose -f deploy/compose/docker-compose.yml --profile quality up
```
