# FastyGo Lab — Progress

**Last updated:** 2026-07-10 (Cycle A foundation)  
**Module:** `github.com/fastygo/lab`  
**CLI:** `lab`

---

## Open cycle

**→ Cycle A — Foundation** (in progress / completing)

Next after A:

1. Cycle B — `quality` L0 (Lighthouse + axe runners)
2. Cycle C — `org` lab (zip lint + Theme Check)
3. Cycle D — `sec` lab (WPScan/nuclei + policy)

---

## Current state

| Area | Status |
|------|--------|
| `.project/` KB | Done (Cycle A) |
| Go domain + contracts + policy | Done (Cycle A) |
| Orchestrator + CLI `demo` lab | Done (Cycle A) |
| Compose profiles stub | Done (Cycle A) |
| Real runners (LH/vnu/wpscan) | Not started |
| WordPress adapter | Not started |
| SaaS API | Not started |

---

## Full path

```
A Foundation ✓
  → B quality L0
    → C org lab
      → D sec lab
        → E static-web adapters
          → F SaaS API + worker
            → G+ runtime / LLM / load labs
```

---

### Cycle A — Foundation

| Stage | Task | Done |
|-------|------|------|
| A0 | `.project/` knowledge base | [x] |
| A1 | Root README, LICENSE, `.gitignore` | [x] |
| A2 | Go module + packages skeleton | [x] |
| A3 | Domain + contracts + policy (TDD) | [x] |
| A4 | Orchestrator + CLI demo lab | [x] |
| A5 | Compose stub + runner contract docs | [x] |

---

### Cycle B — Quality L0 (planned)

| Stage | Task | Done |
|-------|------|------|
| B1 | Lighthouse CI runner image | [ ] |
| B2 | axe / Playwright runner | [ ] |
| B3 | `wordpress` adapter stub | [ ] |
| B4 | Wire `quality` lab manifest | [ ] |

---

### Cycle C — Org lab (planned)

| Stage | Task | Done |
|-------|------|------|
| C1 | Zip lint runner | [ ] |
| C2 | Theme Check container | [ ] |
| C3 | Template URL matrix | [ ] |

---

### Cycle D — Sec lab (planned)

| Stage | Task | Done |
|-------|------|------|
| D1 | WPScan / nuclei runners | [ ] |
| D2 | Policy decisions → baskets | [ ] |
| D3 | Headers / recon checks | [ ] |

---

## Decision log

| Date | Decision |
|------|----------|
| 2026-07 | Separate monorepo `fastygo/lab`; `wpfasty` is client |
| 2026-07 | Go orchestrator; Docker runners; TS only for future SaaS UI |
| 2026-07 | Hexagonal + DDD; checks never in domain |
| 2026-07 | First runnable lab = `demo` (noop); real tools in B+ |
| 2026-07 | CLI binary name `lab`; module `github.com/fastygo/lab` |
| 2026-07 | Labs are pluggable packs — not limited to org/sec/quality |

---

## Commands cheat sheet

```bash
go test ./...
go run ./apps/cli version
go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml

# future
docker compose -f deploy/compose/docker-compose.yml --profile smoke up
```
