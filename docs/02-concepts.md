# 02 — Core concepts

Read this once. Everything else reuses these words.

---

## Manifest

A **Lab Manifest** is a YAML file that declares:

- which **lab** (`org`, `quality`, `sec`, …),
- which **adapter** prepares the target (WordPress, static fixture, static-web),
- ordered **gates**, each with **checks** (runner id + config),
- which **policy pack** classifies findings.

Examples: `testdata/manifests/org.lab.yaml`, `quality-wp.lab.yaml`, `sec.lab.yaml`.

Schema-oriented design notes: [.project/architecture.md](../.project/architecture.md).

---

## Adapter

An **adapter** makes a target available to runners:

| Adapter | Role |
|---------|------|
| `wordpress` | Attach to `baseUrl`, optionally install `themeZip`, seed content |
| `static` | Serve a fixture directory (quality smoke without WP) |
| `static-web` | Preview a Vite/SPA-style root |
| `noop` | Tiny fixture for unit tests / `demo` |

---

## Gate → Check → Runner

| Term | Meaning |
|------|---------|
| **Gate** | Named stage (`C1-zip-lint`, `Q1-lighthouse`, `S2-cve`) |
| **Check** | One item inside a gate (`theme-check`, `axe-wcag`) |
| **Runner** | Implementation — often a Docker image under `runners/` |

Runners should be **swapable**. The orchestrator does not embed Lighthouse or WPScan logic; it launches runners and parses findings.

---

## Finding

A **finding** is one discrete result:

- `code` — stable id (`org.zip.missing_license`, `quality.lighthouse.perf`)
- `severity` — info / low / medium / high / critical (tool-dependent mapping)
- `message`, optional URL / evidence

Many findings are **informational** (proof a URL was checked). Policy decides if they fail the job.

---

## Decision baskets

Policy maps findings into **baskets** so product owners can act:

| Basket (examples) | Meaning |
|-------------------|---------|
| `FIX_THEME` / `CUT_THEME` | Change the theme zip |
| `SITE_DEFAULT_ON` / `SITE_DEFAULT_OFF` | Harden the site baseline (mu-plugin, server, wp-config) |
| `BUDGET` | Threshold / budget trade-off |
| `ACCEPT` | Known OK (e.g. REST index for Gutenberg) |
| `BLOCK_TAG` | Do not claim a wordpress.org tag yet |

Exact names appear in Report JSON and on the dashboard. Treat baskets as the **product language**, not raw WPScan lines.

---

## Report status vs infra error

| Status | Meaning | What you do |
|--------|---------|-------------|
| `pass` | Checks ran; policy green | Ship / celebrate |
| `warn` | Soft issues only | Triage baskets |
| `fail` | Policy fail (high findings, budgets, …) | Fix theme/site — **job still completed** |
| `error` | Orchestrator / runner infra broke | Retry, fix Docker, check logs |

This distinction is critical for juniors: **`fail` is a successful audit with bad news**, not a crashed lab. See [.project/vps/f2-isolation.md](../.project/vps/f2-isolation.md).

---

## Events (timeline)

While a job runs, the orchestrator emits **events** (`run.started`, `gate.started`, `check.finished`, …). The dashboard timeline and SSE stream are built from these. Types are listed in [.project/vps/cycle-f-saas.md](../.project/vps/cycle-f-saas.md).

---

## Preset

A **preset** is a named pointer to a Manifest for the API:

| Preset | Typical use |
|--------|-------------|
| `demo` | Fast smoke without heavy Docker |
| `org` | wordpress.org-oriented gates |
| `quality` | Static fixture quality |
| `quality-wp` | Quality against live WP (logged-out) |
| `sec` | Security lab against owned WP |
| `static-web` | Lighter static-web quality |

`GET /v1/presets` lists them. Bindings (`themeZip`, `baseUrl`, …) override Manifest paths at enqueue time.

---

## Repo layout (mental map)

```text
apps/cli          # local binary
apps/api          # SaaS HTTP + worker queue
apps/web          # Templ dashboard
packages/         # domain, orchestrator, policy, presets, runstore, …
runners/          # Docker entrypoints per tool
deploy/compose/   # WordPress / quality / saas profiles
testdata/         # manifests, fixtures, dist/*.zip
docs/             # you are here
.project/         # architecture + checklists + VPS runbooks
```

Next: [03 — Local quickstart](./03-quickstart-local.md).
