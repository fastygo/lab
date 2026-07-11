# Architecture

## Style

**Hexagonal (ports & adapters) + DDD bounded contexts.**

```text
apps/cli | apps/api (future)
        ↓
packages/orchestrator     # application service
        ↓
packages/domain           # pure model
packages/policy           # Finding → Decision
packages/contracts        # schemas + shared DTOs
        ↓ ports
adapters/*                # TargetAdapter
runners/*                 # Runner (Docker or stub)
```

## Monorepo map

```text
packages/domain/          Lab, Gate, Check, Target, Finding, Decision, Report, Budget
packages/contracts/       JSON Schema + Go types aligned to schemas
packages/policy/          PolicyEngine
packages/orchestrator/    Run(ctx, manifest) → Report
packages/adapters/noop/   Fixture adapter for TDD / demo
apps/cli/                 lab binary
apps/api/                 Cycle F HTTP skeleton (healthz)
packages/runstore/        SaaS job + event persistence ports
runners/                  Container entrypoints
deploy/compose/           Local lab profiles
testdata/manifests/       Example lab manifests
.project/                 Constructor KB
.project/vps/             VPS runbook + Cycle F SaaS roadmap
```

## Module

- Path: `github.com/fastygo/lab`
- Go 1.22+
- Single root `go.mod` (no `go.work` until needed)

## Local vs SaaS

| Mode | Execution |
|------|-----------|
| Local | CLI + Compose profiles (`smoke`, `org`, `sec`, `quality`) |
| SaaS (Cycle F) | API accepts job → queue → worker runs same Manifest → report + **event stream** store |

Same **Lab Manifest** and **Report** schema in both modes. Progress uses `domain.RunEvent` via `ports.EventSink` (see [vps/cycle-f-saas.md](./vps/cycle-f-saas.md)).

```text
apps/cli | apps/api (Cycle F)
        ↓
packages/orchestrator     # Run + EventSink
```

## Scaling

- **Horizontal:** more workers / parallel runners per job
- **Vertical:** new Gate + Runner + optional Adapter; domain stays stable

## Isolation

- Each lab run should use an isolated network/namespace when using Compose
- Security labs: only owned/lab targets (never arbitrary third-party without explicit opt-in)
