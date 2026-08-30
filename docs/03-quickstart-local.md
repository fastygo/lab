# 03 — Local quickstart

Goal: on your machine, run a **demo** job, then optionally one real scenario.

---

## Prerequisites

| Tool | Why | Install |
|------|-----|---------|
| **Go 1.22+** | Build CLI / API / web | [go.dev/dl](https://go.dev/dl/) |
| **Docker** + Compose | Runners + WordPress profiles | [docs.docker.com/get-docker](https://docs.docker.com/get-docker/) |
| **Make** (optional) | Shortcuts in root `Makefile` | OS package / Git Bash on Windows |
| **Git** | Clone the monorepo | [git-scm.com](https://git-scm.com/) |

Optional for theme authors: Bun/Node if you build zips from a theme constructor (outside this repo).

---

## 1. Clone and test

```bash
cd /path/to/Lab
go test ./...
```

If tests pass, the Go module and core packages are healthy.

---

## 2. Fastest win — `demo`

No WordPress required:

```bash
make demo
# or: go run ./apps/cli run -f testdata/manifests/demo.lab.yaml
```

You should see a Report on stdout. This proves Manifest → orchestrator → policy works.

---

## 3. Build runner images (once)

Real org / quality / sec checks need Docker images:

```bash
make runners
```

This builds tags like `lab/lighthouse:local`, `lab/theme-check:local`, `lab/wpscan:local`, …. First build is slow (Chromium, scanners).

---

## 4. Start WordPress (org profile)

For `org`, `quality-wp`, and `sec` against a live site:

```bash
# Set public URL for WP asset links (important on remote hosts)
# Local Docker Desktop example:
export LAB_WP_URL=http://127.0.0.1:8080

make org-up
# waits until http://127.0.0.1:8080 answers 200

make org-seed   # Theme Unit Test import + seed JSON (first time)
```

Compose file: `deploy/compose/docker-compose.yml` (`--profile org`).

WordPress themes & site URLs: [WordPress developer resources](https://developer.wordpress.org/).

---

## 5. Theme zip under test

Manifests default to:

```text
testdata/dist/latte.zip
```

Place your packaged theme there (or pass another path via CLI/API bindings). The zip must be a **theme root** (contains `style.css`), not a monorepo checkout.

Packaging rules overview: [Theme Handbook — Required](https://make.wordpress.org/themes/handbook/review/required/).

---

## 6. Run a product scenario locally

```bash
make org          # Gate 1–4 org checks
make quality-wp   # Lighthouse / axe / vnu / … on WP
make sec          # recon, wpscan, nuclei, theme static, …
```

Expect wall-clock of **many minutes** for quality (Lighthouse median of 3) and sec (scanners).

---

## 7. Local SaaS (API + dashboard)

Terminal A:

```bash
make api
# listens :8090 — LAB_API_URL for workers is loopback by default
```

Terminal B:

```bash
export LAB_API_URL=http://127.0.0.1:8090
make web
# listens :8091 locally (VPS often uses :8092 via LAB_WEB_ADDR)
```

Open http://127.0.0.1:8091/ — empty until you enqueue:

```bash
curl -sS -X POST http://127.0.0.1:8090/v1/runs \
  -H 'Content-Type: application/json' \
  -d '{"preset":"demo","sync":true}'
```

Refresh the dashboard — the run appears.

Optional Postgres for persistence:

```bash
make saas-up
export LAB_DATABASE_URL='postgres://lab:lab@127.0.0.1:5432/lab?sslmode=disable'
```

Without `LAB_DATABASE_URL`, the API uses an in-memory store (lost on restart).

---

## Common pitfalls

| Symptom | Likely cause |
|---------|----------------|
| Theme assets point at wrong host | `LAB_WP_URL` / WP_HOME not set to the URL you open in the browser |
| Theme Check cannot see WP | Org compose not up; wrong `dockerNetwork` / volume names |
| Browser runners fail redirects | Use host-reachable URL (`http://127.0.0.1:8080`), not `http://wordpress` |
| Sec refuses `baseUrl` | `LAB_ALLOWED_BASE_URLS` set and URL not listed |
| Queue “full” / jobs stuck | API restarted mid-run; or worker busy (serial by default) |

More ops detail: [08 — Operations](./08-operations.md).

Next: [04 — Product scenarios](./04-scenarios.md).
