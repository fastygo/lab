# Contracts

Normative shapes live in `packages/contracts/schemas/`. Go types in `packages/contracts` and `packages/domain` must stay aligned.

## Lab Manifest

Identifies which lab to run, target adapter, gates/checks, budgets, and runner bindings.

Required ideas:

- `apiVersion` (e.g. `lab.fastygo.dev/v1`)
- `kind`: `LabManifest`
- `metadata.name`
- `spec.lab` — lab id (`demo`, `org`, …)
- `spec.adapter` — adapter id + config
- `spec.gates[]` — gate id, checks[], optional budgets
- `spec.policy` — optional policy pack id

## Finding

- `id` / `code`
- `gate` / `check`
- `severity`: `critical` | `high` | `medium` | `low` | `info`
- `message`
- `evidence` (optional map/string)
- `target` (optional URL)

## Decision

- `findingCode`
- `basket`: `CUT_TARGET` | `SITE_DEFAULT_ON` | `SITE_DEFAULT_OFF` | `ACCEPT` | `BUDGET` | `FIX_THEME` | `FIX_SITE` | `BLOCK_TAG`
- `rationale`

## Report

- `lab`, `startedAt`, `finishedAt`
- `status`: `pass` | `warn` | `fail`
- `findings[]`, `decisions[]`
- `summary` counts by severity

## Runner I/O (containers)

**Input (env or stdin JSON):** target base URL, check config, budgets.  
**Output (stdout JSON):** array of findings (or `{ "findings": [...] }`).  
Non-zero exit = runner infrastructure failure (not merely “has findings”).

See [runners/README.md](../runners/README.md).
