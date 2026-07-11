# Adapters

An **adapter** prepares and serves a **Target** for lab runners.

## Contract

| Method | Responsibility |
|--------|----------------|
| `Prepare(ctx, config)` | Build/pack artifact (theme zip, `vite build`, `go build`, …) |
| `Serve(ctx)` | Expose HTTP base URL; block until healthy |
| `Matrix(ctx)` | Return URL list for checks |
| `Capabilities()` | e.g. `html`, `http`, `a11y`, `seo`, `wp-theme` |
| `Teardown(ctx)` | Stop server, clean temp dirs |

## Capabilities → labs

Gates declare required capabilities. Orchestrator selects adapters that satisfy them.

| Capability | Used by |
|------------|---------|
| `http` | Most dynamic checks |
| `html` | vnu, axe, SEO parsers |
| `wp-theme` | org / sec WordPress-specific |
| `a11y` | quality ARIA gates |
| `seo` | optional meta/graph profile |

## Built-in

| Adapter | Role |
|---------|------|
| `noop` | Fixture URL + static matrix (Cycle A demo) |
| `wordpress` | Pack/install theme zip in WP (Cycle B+) |
| `static` | Local HTML directory (quality L0 fixture) |
| `static-web` | Vite/React/Vue/Svelte dist or preview (Cycle E) |
| `go-http` | Serve Go binary (later) |
| `php-app` | Generic PHP app (later) |

## Rules

- Adapters know **how to run** a stack, not **how to score** it
- No WP queries or Lighthouse flags inside adapters beyond serve config
- Thin plugins for DX (npm/vite) must not duplicate domain logic
