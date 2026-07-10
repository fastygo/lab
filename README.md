# FastyGo Lab

**Laboratory constructor framework** — Go orchestrator, Docker runners, pluggable lab packs.

Labs: `demo` · `quality` (L0) · `org` · `sec` — extensible for runtime bake-offs, LLM compares, load playgrounds, and more.

## Quick start

```bash
go test ./...
go run ./apps/cli version
go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml
```

### Quality L0 (Lighthouse + vnu + axe)

```bash
make runners                                          # build images
make quality-up                                       # nginx fixture :8091
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml
# Without Docker: report includes runner.docker.unavailable (accepted by lightspeed pack)
```

### Org (zip-lint + Theme Check + HTTP smoke)

```bash
# themeZip defaults to testdata/dist/latte.zip in org.lab.yaml
make org-up && make runners
go run ./apps/cli run -f testdata/manifests/org.lab.yaml
```

### Sec (headers + WPScan)

```bash
go run ./apps/cli run -f testdata/manifests/sec.lab.yaml
# Point wordpress adapter baseUrl at a lab WP; WPScan needs Docker
```

## Layout

```text
packages/domain|contracts|policy|orchestrator|registry
packages/adapters/{noop,static,wordpress}
apps/cli
runners/{lighthouse,axe,theme-check,vnu,wpscan}
deploy/compose/
testdata/manifests|fixtures|dist
.project/
```

## Design

- **Hexagonal + DDD** — checks never live in domain
- **Adapters** prepare/serve targets
- **Runners** wrap tools (in-process or Docker)
- **Policy** maps findings → decisions
- Same **Manifest → Report** for local CLI and future SaaS

## Docs

| Doc | Purpose |
|-----|---------|
| [.project/README.md](.project/README.md) | KB index |
| [.project/PROGRESS.md](.project/PROGRESS.md) | Cycles A–D done; open E |
| [.project/labs/](.project/labs/) | Lab specs |

## License

MIT — see [LICENSE](LICENSE).
