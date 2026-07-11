# FastyGo Lab — project knowledge base

**Purpose:** Single SSOT for agents and humans building the laboratory constructor.  
Read this first, then [PROGRESS.md](./PROGRESS.md), then the open cycle.

**Repo:** `github.com/fastygo/lab` · Go orchestrator + Docker runners  
**Consumers:** `wpfasty` (WordPress themes), later React/Vue/Svelte/Go/PHP adapters

## How to use (agent protocol)

1. Read **Current state** and **Open cycle** in [PROGRESS.md](./PROGRESS.md).
2. Execute only that cycle (or the next open follow-up).
3. After finishing: mark checkboxes, bump **Last updated**.
4. Prefer small PRs per cycle.
5. Discipline: `.project/` = constructor KB · code = `packages/` · runners = containers.

## Map

| Doc | Use |
|-----|-----|
| [PROGRESS.md](./PROGRESS.md) | Cycles, open work, decision log |
| [mental-model.md](./mental-model.md) | Lab / Gate / Runner / Adapter / Finding / Decision |
| [architecture.md](./architecture.md) | Hexagonal + DDD + monorepo + local/SaaS |
| [contracts.md](./contracts.md) | Manifest / Report / Finding schemas |
| [labs.md](./labs.md) | Lab catalog + extension rules |
| [adapters.md](./adapters.md) | Runtime adapter contract |
| [policy.md](./policy.md) | Decision baskets |
| [anti-patterns.md](./anti-patterns.md) | Never regress |
| [SOURCES.md](./SOURCES.md) | External tools and handbooks |
| [constructor/lab-constructor.md](./constructor/lab-constructor.md) | Add a new lab pack |
| [constructor/compose-profiles.md](./constructor/compose-profiles.md) | L0–L3 Compose profiles |
| [labs/org.md](./labs/org.md) | WordPress.org readiness lab |
| [labs/sec.md](./labs/sec.md) | Security / hardening lab |
| [labs/quality.md](./labs/quality.md) | Lighthouse / W3C / ARIA / SEO lab |
| [check/](./check/) | Product checklists (theme / validate / security) |
| [check/audit-progress.md](./check/audit-progress.md) | **Full check coverage** `[x]` / `[~]` / `[ ]` |
| [vps/README.md](./vps/README.md) | **Clean VPS from source** — bootstrap runbook (EN) |

## Hard rules

- Checks live in **runners**, not domain
- Runtime-specific knowledge lives in **adapters**
- `wpfasty` is a **client**, not owner of this repo
- Security labs only against owned/lab targets
- Framework first: org/sec/quality are the first labs, not the only ones
