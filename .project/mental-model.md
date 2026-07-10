# Mental model

## One sentence

A **Lab** is a named pack of **Gates**; each Gate runs **Checks** via **Runners** against a **Target** prepared by an **Adapter**; results are **Findings**, mapped by **Policy** into **Decisions**, merged into a **Report**.

## Terms

| Term | Meaning |
|------|---------|
| **Lab** | Product-facing pack (`demo`, `org`, `sec`, `quality`, later `llm-bakeoff`, …) |
| **Gate** | Named stage inside a lab (e.g. `Q1-lighthouse`, `S1-recon`) |
| **Check** | Single assertion or tool invocation inside a gate |
| **Runner** | Container (or in-process stub) that executes a check and emits findings JSON |
| **Adapter** | Runtime-specific prepare/serve/matrix (`wordpress`, `react`, `noop`) |
| **Target** | Running system under test (base URL + metadata) |
| **Finding** | One issue or observation (severity, code, message, evidence) |
| **Decision** | Policy basket for a finding (`CUT_TARGET`, `SITE_DEFAULT_ON`, …) |
| **Report** | Aggregated findings + decisions + pass/warn/fail |
| **Budget** | Numeric thresholds (Lighthouse scores, byte weights) |
| **Manifest** | YAML/JSON describing which lab/gates/adapters/runners to run |

## Flow

```text
Manifest → Adapter.Prepare/Serve → Runner(s) per Check → Findings
         → Policy.Map → Decisions → Report → exit code
```

## What belongs where

| Layer | Owns | Must not own |
|-------|------|--------------|
| Domain | Entities, invariants, merge rules | Docker, WP APIs, Lighthouse |
| Policy | Finding → Decision | Tool CLIs |
| Orchestrator | Wiring ports, order of gates | Tool-specific flags |
| Adapter | How to build/serve a runtime | Scoring rules |
| Runner | How to invoke one tool | Product policy baskets |
