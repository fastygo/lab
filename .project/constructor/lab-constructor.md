# Lab constructor — add a new lab

End-to-end checklist for a new lab pack.

## Steps

1. **Spec** — write `.project/labs/<id>.md` (purpose, gates, runners, adapters, budgets, out of scope).
2. **Catalog** — add row to [labs.md](../labs.md).
3. **Schema** — ensure Manifest still validates (`packages/contracts/schemas`).
4. **Manifest** — `testdata/manifests/<id>.lab.yaml` example.
5. **Adapter** — reuse existing or add `packages/adapters/<name>` implementing the port.
6. **Runners** — add `runners/<tool>/` Dockerfile + README; emit findings JSON.
7. **Policy** — map new finding codes in a policy pack if needed.
8. **Compose** — add/extend profile in `deploy/compose` ([compose-profiles.md](./compose-profiles.md)).
9. **CLI** — register lab id if using a registry; or load purely from Manifest.
10. **Tests** — domain/policy unit tests + one e2e smoke with stub or Compose L0.
11. **PROGRESS** — new cycle or stages; bump Last updated.

## Definition of done

- `lab run -f testdata/manifests/<id>.lab.yaml` produces a Report
- Documented capabilities and required runners
- No domain imports of the tool SDK
