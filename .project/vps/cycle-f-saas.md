# Cycle F — SaaS API / workers / dashboard

Roadmap + checklists for turning local Lab CLI into a **job-based SaaS** that runs the same three product scenarios (`org`, `quality`, `sec`) with progress events, reports, schedules, and Telegram/Slack notify.

Host context: this lives next to the [VPS runbook](./README.md) because cloud workers will reuse the same Docker images and compose profiles on lab hosts.

**Status:** 2026-07-11 — Cycle F core done (F0–F4). Optional: F3.5 compare, richer export.

---

## Goal

| Outcome | Definition of done |
|---------|-------------------|
| Three scenarios @ 100% | Presets `org` / `quality` / `sec` run as **jobs** on workers (same Manifest/Report as CLI) |
| Process visibility | Thin **event stream** per run (gate/check lifecycle) → dashboard timeline |
| Reporting | Persist Report JSON; list/filter runs; severity + decision baskets |
| Notify | On finish: Telegram and/or Slack; optional schedules (cron) |

Same contracts as local mode ([architecture.md](../architecture.md)):

```text
Manifest → Adapter → Runners → Findings → Policy → Report
                 ↘ EventSink (progress) ↗
```

---

## Architecture (target)

```text
┌─────────────┐     POST /v1/runs      ┌──────────┐
│  Dashboard  │ ─────────────────────► │ apps/api │
│  (F0–F2)    │ ◄── SSE / GET report ──│          │
└─────────────┘                        └────┬─────┘
                                            │
                     ┌──────────────────────┼──────────────────────┐
                     ▼                      ▼                      ▼
              Postgres                 Queue                  Object store
           runs/events/              (Redis/NATS)            artifacts +
           schedules/notify                                  reports.json
                     ▲                      │
                     │                      ▼
                     │               ┌─────────────┐
                     └───────────────│   Worker    │  = orchestrator.Run
                                     │ + Docker    │    + EventSink → API/DB
                                     └─────────────┘
                                            │
                                            ▼
                                     Notifier (TG/Slack)
```

**Hard rules**

- API does **not** reimplement Lighthouse/WPScan — only jobs + store + notify.
- Sec jobs: **owned/lab targets only** (explicit opt-in for anything else).
- Theme zip never carries login-limiter / webhook secrets.

---

## Thin event layer (foundation)

Events are the bridge from CLI today → dashboard tomorrow. CLI may ignore them; API/worker must persist them.

### Event types

| `type` | When | Payload (typical) |
|--------|------|-------------------|
| `run.started` | Engine begins | `lab`, `runId?` |
| `adapter.ready` | After Serve | `adapter`, `baseUrl` |
| `gate.started` | Before checks in gate | `gate` |
| `check.started` | Before runner | `gate`, `check`, `runner` |
| `check.finished` | After runner | `gate`, `check`, `runner`, `findingCount`, `durationMs` |
| `gate.finished` | After all checks in gate | `gate`, `findingCount` |
| `run.finished` | After Report | `status`, `summary`, `durationMs` |
| `run.failed` | Fatal orchestrator error | `error` |

### Port

```go
// packages/orchestrator/ports
type EventSink interface {
    Emit(ctx context.Context, ev domain.RunEvent) error
}
```

Nop sink for CLI default; memory sink for tests; HTTP/DB sink for SaaS worker.

### Checklist — event layer

- [x] `domain.RunEvent` (+ JSON tags)
- [x] `ports.EventSink`
- [x] Orchestrator emits lifecycle events
- [x] `memory.EventSink` for tests
- [x] Unit test: demo run produces ordered events
- [x] JSON Schema `run-event.schema.json`
- [ ] CLI flag `--events` (optional stderr/NDJSON) — nice-to-have
- [x] Worker sink → Postgres `run_events` when `LAB_DATABASE_URL` / `DATABASE_URL` set (`packages/runstore/postgres`)
- [x] `GET /v1/runs/{id}/events` (JSON; SSE later in F3)

---

## Phased roadmap

### F0 — Foundation (done)

**Intent:** domain + orchestrator ready for SaaS without shipping HTTP yet.

| # | Item | Status |
|---|------|--------|
| F0.1 | Event types + EventSink + engine wire-up | [x] |
| F0.2 | Contract schema for RunEvent | [x] |
| F0.3 | Spec docs: this file + `.project/labs/saas.md` | [x] |
| F0.4 | Skeleton `apps/api` (healthz only) | [x] |
| F0.5 | Skeleton `packages/runstore` interface (memory) | [x] |
| F0.6 | Compose profile note `saas` (worker shape) | [x] |

**Exit:** `go test` green; demo run records ≥ `run.started` … `run.finished`.

---

### F1 — API + worker MVP (quality first)

| # | Item | Status |
|---|------|--------|
| F1.1 | Postgres schema: `runs`, `run_events`, `artifacts` + `DATABASE_URL` / `LAB_DATABASE_URL` | [x] |
| F1.2 | `POST /v1/runs` — body: `{ "preset"|"manifestPath", "sync?" }` | [x] |
| F1.3 | `GET /v1/runs/{id}` — status + progress summary | [x] |
| F1.4 | `GET /v1/runs/{id}/report` — Report JSON | [x] |
| F1.5 | `GET /v1/runs` — list/filter | [x] |
| F1.6 | Worker process: dequeue → `orchestrator.Run` → save report + events | [x] (in-process) |
| F1.7 | Preset `quality` (static fixture) E2E on VPS worker | [x] (API job; fixture may `fail` on budgets) |
| F1.8 | Artifact upload or `themeZip` / fixture path binding | [x] (`themeZip`/`root`/`baseUrl`/`config`/`checkConfig` on POST) |

**Exit:** Dashboard-less curl can create a **demo** or **quality** run and fetch a Report.

**VPS checklist (F1)**

- [x] Worker host has `make runners` images used by quality
- [~] API + worker + Postgres (compose or managed) — API+worker verified; Postgres URI ready, memory used on smoke
- [x] `POST` quality → `status=pass|fail|warn` + report in store

### POST bindings (F1.8)

```json
{
  "preset": "org",
  "themeZip": "testdata/dist/latte.zip",
  "baseUrl": "http://127.0.0.1:8080",
  "root": "testdata/fixtures/quality-site",
  "config": { "seedFile": "testdata/fixtures/org-seed.json" },
  "checkConfig": { "dockerNetwork": "fastygo-lab_lab" },
  "sync": false
}
```

Paths are snapshotted into the run manifest; adapters resolve relative to `LAB_REPO_ROOT`.
---

### F2 — Three scenarios @ 100%

| # | Item | Status |
|---|------|--------|
| F2.1 | Preset `quality-wp` (live WP logged-out) | [x] (VPS API) |
| F2.2 | Preset `org` — compose org profile + zip + seed on worker | [x] (VPS API) |
| F2.3 | Preset `sec` — owned WP only; env for `LAB_WP_*` / tokens | [x] (VPS API + `LAB_ALLOWED_BASE_URLS`) |
| F2.4 | Preset `static-web` (optional fourth) | [x] (VPS API, pass) |
| F2.5 | Isolation: per-run network / cleanup policy documented | [x] ([f2-isolation.md](./f2-isolation.md)) |
| F2.6 | Golden “known fail” fixtures vs exit codes documented | [x] ([f2-isolation.md](./f2-isolation.md)) |

**Exit:** All three product labs runnable as jobs with persisted reports matching CLI shape.

**VPS checklist (F2)**

- [x] Org profile up; `LAB_WP_URL` public
- [x] Sec runners built (`wpscan`, `nuclei`, `phpcs-security`, `semgrep`, …)
- [x] Job timeout + disk budget per run (documented; 60 min worker ctx)
- [x] Deny arbitrary third-party URLs — `LAB_ALLOWED_BASE_URLS` in wordpress adapter
---

### F3 — Dashboard (process + reporting)

| # | Item | Status |
|---|------|--------|
| F3.0 | **F0 UI** — table of runs (lab, status, started, duration, link) | [x] (`apps/web`, ui8kit templ) |
| F3.1 | Run detail — Report summary + findings table | [x] |
| F3.2 | Decision baskets breakdown | [x] |
| F3.3 | Timeline from `run_events` (gate → check) | [x] |
| F3.4 | Live SSE while `running` | [x] (`/v1/runs/{id}/events/stream` + dashboard EventSource) |
| F3.5 | Compare two runs (regression hint) | [ ] |
| F3.6 | Export markdown / download JSON | [~] (link to report JSON) |

**Exit:** Operator can watch a run’s checks and open a shareable report without SSH.

```bash
make api   # :8090
make web   # :8091  LAB_API_URL=http://127.0.0.1:8090
# open http://127.0.0.1:8091/
```

UI primitives: vendored `third_party/ui8kit` (from `@ui8kit/codegen` generated Templ).
---

### F4 — Schedules + Telegram / Slack

| # | Item | Status |
|---|------|--------|
| F4.1 | Table `schedules` (cron, lab preset, notify channels, enabled) | [x] |
| F4.2 | Scheduler tick → enqueue run | [x] (`LAB_SCHEDULER_INTERVAL`, default 30s) |
| F4.3 | Slack incoming webhook notifier on `run.finished` | [x] (`SLACK_WEBHOOK_URL`) |
| F4.4 | Telegram bot notifier (`sendMessage`) | [x] (`TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID`) |
| F4.5 | Notify filters: `always` \| `fail` \| `warn+fail` | [x] (`LAB_NOTIFY_ON`) |
| F4.6 | `POST /v1/notify/test` | [x] |
| F4.7 | Secrets via env / vault — never in Manifest committed to git | [x] |

**Message sketch**

```text
[sec] FAIL — high:4 medium:7 (27 findings)
gates: S1✗ S2✓ S3✗ S4~
→ https://lab.example/runs/<id>
```

**VPS checklist (F4)**

- [x] `SLACK_WEBHOOK_URL` and/or `TELEGRAM_BOT_TOKEN` + `TELEGRAM_CHAT_ID` on worker/API
- [x] Scheduler API: `POST /v1/schedules` + tick enqueue (e.g. weekly quality)
- [x] Alert filter via `LAB_NOTIFY_ON` (e.g. `fail` for sec nightly)

### Schedules API

```bash
# every hour on the hour — quality preset
curl -sS -X POST http://127.0.0.1:8090/v1/schedules \
  -H 'Content-Type: application/json' \
  -d '{"cron":"0 * * * *","preset":"quality","enabled":true}'

# fire now
curl -sS -X POST http://127.0.0.1:8090/v1/schedules/<id>/fire
```
---

## Suggested build order (start now)

```text
1. EventSink in orchestrator          ← done
2. runstore memory + apps/api + worker MVP ← done
3. Postgres via DATABASE_URL / LAB_DATABASE_URL (pgx) ← done
4. quality preset E2E on workers      ← done (VPS API)
5. org + sec presets on VPS workers   ← done (F2)
6. Dashboard F0 → timeline + SSE     ← done (F3)
7. Schedules + Slack/Telegram         ← done (F4)
```

### Connection URL (Postgres / Supabase)

Same **connection string / URI** pattern as Supabase «Database URL» — not a separate product extension:

```text
postgres://USER:PASSWORD@HOST:5432/DBNAME?sslmode=require
```

Env: `LAB_DATABASE_URL` or `DATABASE_URL`. Driver: **pgx** (`packages/runstore/postgres`).  
Local: `docker compose … --profile saas up -d postgres` → `postgres://lab:lab@127.0.0.1:5432/lab?sslmode=disable`.  
Without URL the API keeps the in-memory store.

---

## Data model (draft)

### `runs`

| Column | Notes |
|--------|-------|
| `id` | UUID |
| `lab` | `org` \| `quality` \| `sec` \| … |
| `status` | `queued` \| `running` \| `pass` \| `warn` \| `fail` \| `error` |
| `manifest_json` | Snapshot |
| `report_json` | Final Report |
| `started_at` / `finished_at` | |
| `error` | Fatal message |

### `run_events`

| Column | Notes |
|--------|-------|
| `id` | bigserial |
| `run_id` | FK |
| `ts` | timestamptz |
| `type` | event type string |
| `payload` | jsonb |

### `schedules`

| Column | Notes |
|--------|-------|
| `id` | UUID |
| `cron` | standard 5-field |
| `lab_preset` | |
| `notify` | jsonb `{on, channels}` |
| `enabled` | bool |

---

## Presets map (CLI → SaaS)

| Preset | Manifest | Worker needs |
|--------|----------|--------------|
| `quality` | `testdata/manifests/quality.lab.yaml` | LH/axe/vnu/css/extras images |
| `quality-wp` | `quality-wp.lab.yaml` | WP :8080 + host network |
| `org` | `org.lab.yaml` | org compose + theme zip + seed |
| `sec` | `sec.lab.yaml` | WP + sec runners; owned URL only |
| `static-web` | `quality-staticweb.lab.yaml` | lighter quality runners |

---

## Risks

| Risk | Mitigation |
|------|------------|
| Long sec/org jobs block workers | Queue + concurrency limit; separate pools |
| Docker Hub 429 on workers | Pre-pull images; mirrors ([VPS README](./README.md)) |
| Dashboard expects live progress | EventSink from day one — do not wait for UI |
| Notify spam | Default `on: [fail, warn]`; digest later |
| Multi-tenant auth | Out of F0–F2; single-tenant lab token first |

---

## Definition of “Cycle F done”

- [x] F0 event layer merged and tested
- [x] F1 quality job via API on VPS
- [x] F2 org + sec jobs via API
- [x] F3.0–F3.3 dashboard (list + report + timeline)
- [x] F3.4 SSE live timeline
- [x] F4 Slack/Telegram + schedules

Nice-to-have after Cycle F: multi-tenant, billing, fancy compare UI (F3.5+).

---

## Related docs

- [VPS clean-host runbook](./README.md)
- [Architecture local vs SaaS](../architecture.md)
- [Mental model](../mental-model.md)
- [Contracts](../contracts.md)
- [Compose profiles](../constructor/compose-profiles.md) (`saas` profile stub)
- Product checklists: [audit-progress.md](../check/audit-progress.md)
