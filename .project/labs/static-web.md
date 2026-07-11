# Lab: static-web — Cycle E SPA / Vite targets

**Id:** `static-web`  
**Cycle:** E  
**Adapter:** `static-web`

## Purpose

Serve React / Vue / Svelte / Vite **build output** (or live preview) so existing quality runners (Lighthouse, axe, vnu, SEO) can score non-WordPress frontends without forking the orchestrator.

## Adapter modes

| Mode | Needs Node? | Behavior |
|------|-------------|----------|
| `dist` (default) | No | Serve `dist/` with SPA `index.html` fallback |
| `preview` | Yes | Run `previewCommand` (default: `npx vite preview`) |
| `baseUrl` set | No | Use an already-running preview (Compose / CI) |

## Config keys

| Key | Default | Notes |
|-----|---------|-------|
| `root` | `testdata/fixtures/static-web-app` | App / monorepo package dir |
| `dist` | `dist` | Build output relative to root |
| `mode` | `dist` | `dist` \| `preview` |
| `build` | `false` | Run `buildCommand` in Prepare |
| `buildCommand` | `npm run build` | |
| `previewCommand` | vite preview… | `{port}` substituted |
| `spa` | `true` | Client-route fallback to `index.html` |
| `framework` | `vite` | Metadata only (`react`/`vue`/`svelte`/…) |
| `matrix` | `/,/about,/about.html` | CSV paths |
| `baseUrl` | — | External target URL |

## Manifest

`testdata/manifests/quality-staticweb.lab.yaml` — SEO + vnu + axe against the fixture.

```bash
go run ./apps/cli run -f testdata/manifests/quality-staticweb.lab.yaml
```

## Out of scope

- Framework unit tests / Storybook
- SSR hydrate correctness (separate bakeoff lab later)
- SaaS multi-tenant routing
