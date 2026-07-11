# FastyGo Lab — Progress

**Last updated:** 2026-07-11 (Org Gate 3 Unit Test XML + attachment + notice hunter)  
**Module:** `github.com/fastygo/lab`  
**CLI:** `lab` v0.2.0

**Check coverage (full item list):** [check/audit-progress.md](./check/audit-progress.md) — mark `[x]`/`[~]`/`[ ]` there so nothing is forgotten.

---

## Open cycle

**→ Next product checks** (see [check/audit-progress.md](./check/audit-progress.md) “Suggested next”)

1. Org Gate 4 — Playwright keyboard (skip / nav / sheet / search)  
2. Quality Q3 — stylelint  
3. Sec S1 — user enum + sensitive files + REST  
4. Cycle E — static-web adapters (React/Vue/Svelte)

---

## Current state

| Area | Status |
|------|--------|
| `.project/` KB | Done |
| Go domain + contracts + policy | Done |
| Orchestrator + CLI | Done (v0.2.0) |
| Docker runner port | Done (Cycle B; zip mount + compose net) |
| `quality` L0 (LH + axe + vnu + static) | Done |
| `wordpress` adapter stub | Done |
| `org` Gate 1 zip-lint + Gate 2 Theme Check + Gate 3 HTTP smoke | Done |
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
| B5 | Q2 vnu runner + quality gate | [x] |

---

### Cycle C — Org lab

| Stage | Task | Done |
|-------|------|------|
| C1 | In-process zip-lint + fixtures/tests | [x] |
| C1+ | Gate 1 full zip-lint (ext, Resources, policy, min twins, slug) | [x] |
| C2 | Compose `org` + Theme Check headless (zip install) | [x] |
| C3 | HTTP smoke asserts on URL matrix + wordpress-org policy | [x] |
| C3+ | Theme Unit Test XML seed + attachment URL + notice hunter | [x] |

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
| 2026-07 | Gate 2: theme zip mounted into theme-check container; share compose WP volume/network |
| 2026-07 | Gate 3: `http-matrix` performs real GET asserts (not list-only) |
| 2026-07-11 | Gate 3+: Unit Test XML in fixtures; seed from theme-check; notice-hunter Docker runner |

---

## Commands cheat sheet

```bash
go test ./...
go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml   # needs Docker images for real LH/axe/vnu
go run ./apps/cli run -f testdata/manifests/org.lab.yaml       # themeZip=testdata/dist/latte.zip
go run ./apps/cli run -f testdata/manifests/sec.lab.yaml

make runners   # docker build lab/*:local (incl. theme-check, vnu, notice-hunter)
make org-up    # compose --profile org
make org-seed  # import Theme Unit Test XML + write org-seed.json
make quality-up
```
