# FastyGo Lab — Progress

**Last updated:** 2026-07-10 (Cycles B–D)  
**Module:** `github.com/fastygo/lab`  
**CLI:** `lab` v0.2.0

---

## Open cycle

**→ Cycle E — static-web adapters** (React/Vue/Svelte)

Suggested next:

1. Cycle E — `static-web` adapter + framework presets
2. Cycle F — SaaS API + worker
3. Quality L1 — vnu / stylelint / Q2–Q6

---

## Current state

| Area | Status |
|------|--------|
| `.project/` KB | Done |
| Go domain + contracts + policy | Done |
| Orchestrator + CLI | Done (v0.2.0) |
| Docker runner port | Done (Cycle B) |
| `quality` L0 (LH + axe + static) | Done |
| `wordpress` adapter stub | Done |
| `org` zip-lint + theme-check compose | Done |
| `sec` headers + wpscan | Done |
| SaaS API | Not started |

---

## Full path

```
A Foundation ✓
  → B quality L0 ✓
    → C org lab ✓
      → D sec lab ✓
        → E static-web adapters
          → F SaaS API + worker
            → G+ runtime / LLM / load labs
```

---

### Cycle A — Foundation

| Stage | Task | Done |
|-------|------|------|
| A0–A5 | KB, skeleton, demo lab, compose stub | [x] |

---

### Cycle B — Quality L0

| Stage | Task | Done |
|-------|------|------|
| B0 | Docker Runner + injectable exec tests; CLI registry | [x] |
| B1 | Lighthouse runner image | [x] |
| B2 | axe runner image | [x] |
| B3 | `static` + `wordpress` stub adapters | [x] |
| B4 | `quality.lab.yaml` + lightspeed policy | [x] |

---

### Cycle C — Org lab

| Stage | Task | Done |
|-------|------|------|
| C1 | In-process zip-lint + fixtures/tests | [x] |
| C1+ | Gate 1 full zip-lint (ext, Resources, policy, min twins, slug) | [x] |
| C2 | Compose `org` + theme-check runner | [x] |
| C3 | `org.lab.yaml` + http-matrix + wordpress-org policy | [x] |

---

### Cycle D — Sec lab

| Stage | Task | Done |
|-------|------|------|
| D1 | Headers / recon runner | [x] |
| D2 | WPScan Docker runner | [x] |
| D3 | secure-baseline policy + `sec.lab.yaml` | [x] |

---

## Decision log

| Date | Decision |
|------|----------|
| 2026-07 | Separate monorepo `fastygo/lab`; `wpfasty` is client |
| 2026-07 | Go orchestrator; Docker runners; checks never in domain |
| 2026-07 | First runnable lab = `demo`; real tools in B+ |
| 2026-07 | Quality L0 uses static fixture; WP stub for later compose |
| 2026-07 | Missing Docker → finding `runner.docker.unavailable` (no panic) |
| 2026-07 | Zip-lint Gate 1 P0: full packaging checklist from `.project/check/theme-check.md` |
| 2026-07 | Policy packs: default, lightspeed, wordpress-org, secure-baseline |

---

## Commands cheat sheet

```bash
go test ./...
go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml   # needs Docker images for real LH/axe
go run ./apps/cli run -f testdata/manifests/org.lab.yaml       # zip-lint needs themeZip in adapter config
go run ./apps/cli run -f testdata/manifests/sec.lab.yaml

make runners   # docker build lab/*:local
docker compose -f deploy/compose/docker-compose.yml --profile quality up -d
docker compose -f deploy/compose/docker-compose.yml --profile org up -d
```
