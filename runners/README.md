# Runners

A **runner** executes one tool (or a small tool family) and emits **findings** JSON.

## Contract

### Input

Prefer env vars + optional stdin JSON:

| Key | Meaning |
|-----|---------|
| `LAB_TARGET_URL` | Base URL of the target |
| `LAB_CHECK_ID` | Check id from the manifest |
| `LAB_GATE_ID` | Gate id |
| `LAB_CONFIG_JSON` | Check-specific config |

### Output

Stdout: either a JSON array of findings or:

```json
{ "findings": [ { "code": "...", "severity": "info", "message": "..." } ] }
```

Stderr: logs only.  
Exit `0` = runner completed (findings may still be non-empty).  
Non-zero = infrastructure failure.

## Layout (later cycles)

```text
runners/
  lighthouse/
  vnu/
  axe/
  wpscan/
  theme-check/
  zip-lint/
  stub/          # optional containerized echo runner
```

Cycle A uses an **in-process stub runner** in Go for the `demo` lab — no images required.
