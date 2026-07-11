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

F0 done. F1 MVP: create/list/get run, report, events; sync or queue; demo preset green in tests. Postgres + quality E2E still open.
