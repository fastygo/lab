# Lab: saas — Cycle F job API (draft)

**Id:** `saas` (platform, not a Manifest lab pack)  
**Cycle:** F  
**Spec SSOT:** [../vps/cycle-f-saas.md](../vps/cycle-f-saas.md)

## Purpose

Run existing labs (`org`, `quality`, `sec`, …) as **jobs** with:

- persisted Reports
- thin **progress events** for dashboards
- optional schedules + Telegram/Slack

## Non-goals (Cycle F)

- Reimplementing runners inside the API
- Multi-tenant billing
- Scanning arbitrary third-party sites

## Components

| Piece | Path |
|-------|------|
| Event port | `packages/orchestrator/ports.EventSink` |
| API | `apps/api` (`make api`, `:8090`) |
| Worker | in-process in `apps/api` + `packages/worker` |
| Presets | `packages/presets` |
| Store | `packages/runstore` (+ `memory`) |

## Status

F0–F2 done. Product presets runnable as API jobs on VPS (`org`, `quality`, `quality-wp`, `sec`, `static-web`).  
`LAB_ALLOWED_BASE_URLS` guards owned targets. Isolation notes: [../vps/f2-isolation.md](../vps/f2-isolation.md). Next: F3 dashboard.
