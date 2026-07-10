# FastyGo Lab

**Laboratory constructor framework** — Go orchestrator, Docker runners, pluggable lab packs.

Not a fixed trio of checks. Near-term labs: WordPress.org readiness (`org`), security (`sec`), quality (`quality`). Later: runtime bake-offs, LLM compares, cloud load playgrounds, and more.

## Status

**Cycle A — Foundation:** domain, contracts, policy, orchestrator, CLI `demo` lab, Compose stubs.  
See [.project/PROGRESS.md](.project/PROGRESS.md).

## Quick start

```bash
go test ./...
go run ./apps/cli version
go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml
```

## Layout

```text
packages/domain|contracts|policy|orchestrator
packages/adapters/noop
apps/cli                 # lab binary
runners/                 # container contracts (tools in later cycles)
deploy/compose/          # local lab profiles
.project/                # knowledge base for agents & humans
testdata/manifests/
```

## Design

- **Hexagonal + DDD** — checks never live in domain
- **Adapters** prepare/serve targets (WordPress, static web, …)
- **Runners** wrap real tools (Lighthouse, axe, WPScan, …)
- **Policy** maps findings → decisions (`CUT_TARGET`, `SITE_DEFAULT_ON`, …)
- Same **Manifest → Report** for local CLI and future SaaS workers

## Docs

| Doc | Purpose |
|-----|---------|
| [.project/README.md](.project/README.md) | KB index |
| [.project/architecture.md](.project/architecture.md) | Architecture |
| [.project/labs.md](.project/labs.md) | Lab catalog |

## License

MIT — see [LICENSE](LICENSE).
