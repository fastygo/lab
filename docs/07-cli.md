# 07 — CLI cheat sheet

The CLI is the fastest feedback loop while developing runners or Manifests.

```bash
go run ./apps/cli <command>
# after: go build -o lab ./apps/cli  →  ./lab …
```

Makefile wrappers call the same binaries.

---

## Common commands

| Command | What |
|---------|------|
| `make test` | `go test ./...` |
| `make demo` | Run demo Manifest |
| `make org` / `quality` / `quality-wp` / `sec` / `static-web` | Product Manifests |
| `make runners` | Build all local runner images |
| `make org-up` | Start org WordPress profile |
| `make org-seed` | Import Theme Unit Test + write `org-seed.json` |
| `make quality-up` | Quality compose profile (static fixture path) |
| `make api` / `make web` | Local SaaS processes |
| `make saas-up` | Postgres for SaaS profile |

Pass extra CLI args:

```bash
make cli ARGS='labs'
make cli ARGS='run -f testdata/manifests/org.lab.yaml'
```

---

## Manifest-first workflow

1. Copy an existing `testdata/manifests/*.lab.yaml`.  
2. Point `themeZip` / `baseUrl` / runner config at your target.  
3. `go run ./apps/cli run -f path/to/yours.lab.yaml`.  
4. Promote to a **preset** in `packages/presets` only when the API should know the name.

---

## Exit codes

Align with Report status: infra `error` vs policy `fail`/`warn`/`pass`. Automation should retry **errors**, not every **fail**.

Next: [08 — Operations](./08-operations.md).
