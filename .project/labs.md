# Labs catalog

## Built-in / planned

| Lab id | Purpose | Cycle |
|--------|---------|-------|
| `demo` | Framework smoke (noop adapter + stub runner) | A |
| `org` | WordPress.org theme readiness | C |
| `sec` | Attack surface + hardening decisions | D |
| `quality` | Lighthouse ×4, HTML5, CSS, ARIA, SEO | B |
| `static-web` | Vite/SPA targets via `static-web` adapter | E |

## Future (examples — not scheduled)

| Lab id | Purpose |
|--------|---------|
| `runtime-bakeoff` | Compare Go / PHP / React / Vue / Svelte targets |
| `llm-compare` | Score LLM outputs against fixtures |
| `load-cloud` | Load-test playground across cloud providers |

## Extension rules

1. Add a lab **spec** under `.project/labs/<id>.md`.
2. Add a **manifest example** under `testdata/manifests/`.
3. Register gates/checks; bind **runners** (new image if needed).
4. Reuse adapters via `capabilities` — do not fork orchestrator.
5. Update [PROGRESS.md](./PROGRESS.md) with a new cycle letter if large.
6. Never put tool CLIs inside `packages/domain`.

See [constructor/lab-constructor.md](./constructor/lab-constructor.md).
