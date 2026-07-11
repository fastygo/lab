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

| Piece | Path (planned) |
|-------|----------------|
| Event port | `packages/orchestrator/ports.EventSink` |
| API | `apps/api` |
| Worker | `apps/worker` or `apps/api` worker mode |
| Store | `packages/runstore` |

## Status

F0 event layer in orchestrator — started.
