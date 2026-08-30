# 05 — API & dashboard

Cycle F turned the CLI engine into a **job API** plus a thin **dashboard**. You enqueue work with HTTP; you watch progress in the browser.

Architecture roadmap: [.project/vps/cycle-f-saas.md](../.project/vps/cycle-f-saas.md).

---

## Surfaces

| Service | Default local | Example VPS |
|---------|---------------|-------------|
| `lab-api` | `:8090` | http://HOST:8090 |
| `lab-web` | `:8091` | http://HOST:8092 |

- **Internal** SSR calls use `LAB_API_URL` (often `http://127.0.0.1:8090`).  
- **Browser** links (healthz, report JSON/MD) use `LAB_API_PUBLIC_URL`, or derive from the request Host.

---

## Dashboard (what it is / is not)

**Is:** run list, run detail (summary, findings, baskets), live timeline (SSE), compare view.

**Is not (yet):** a “Create run” form. The empty state tells you to `POST /v1/runs`. Treat the UI as an **ops console**, not a full admin SPA.

Open the dashboard → refresh after enqueue → click a run → watch Timeline while status is `running`.

---

## Enqueue a run

```http
POST /v1/runs
Content-Type: application/json
```

```json
{
  "preset": "org",
  "themeZip": "testdata/dist/latte.zip",
  "baseUrl": "http://127.0.0.1:8080",
  "sync": false
}
```

| Field | Purpose |
|-------|---------|
| `preset` | Named Manifest (`org`, `quality-wp`, `sec`, `demo`, …) |
| `manifestPath` | Alternative to preset (path under repo) |
| `themeZip` | Override zip path (relative to `LAB_REPO_ROOT`) |
| `baseUrl` | WordPress (or site) URL for adapters/runners |
| `root` | Static / static-web root |
| `config` / `checkConfig` | Extra adapter / check overrides |
| `sync` | `true` = wait for completion in the HTTP response (handy for demos; avoid for long LH/sec) |

Response includes `id` and `status` (`queued` → `running` → `pass|fail|warn|error`).

### Examples

```bash
API=http://127.0.0.1:8090   # or public host:8090

# Smoke
curl -sS -X POST "$API/v1/runs" -H 'Content-Type: application/json' \
  -d '{"preset":"demo","sync":true}'

# Org (async)
curl -sS -X POST "$API/v1/runs" -H 'Content-Type: application/json' \
  -d '{"preset":"org","themeZip":"testdata/dist/latte.zip","baseUrl":"http://127.0.0.1:8080"}'

# Quality on WP
curl -sS -X POST "$API/v1/runs" -H 'Content-Type: application/json' \
  -d '{"preset":"quality-wp","themeZip":"testdata/dist/latte.zip","baseUrl":"http://127.0.0.1:8080"}'

# Security (owned URL only)
curl -sS -X POST "$API/v1/runs" -H 'Content-Type: application/json' \
  -d '{"preset":"sec","themeZip":"testdata/dist/latte.zip","baseUrl":"http://127.0.0.1:8080"}'
```

List presets: `GET /v1/presets`.

---

## Read results

| Method | Path | Returns |
|--------|------|---------|
| GET | `/v1/runs` | List (`?lab=&limit=`) |
| GET | `/v1/runs/{id}` | Status + summary |
| GET | `/v1/runs/{id}/report` | Full Report JSON |
| GET | `/v1/runs/{id}/report.md` | Markdown report |
| GET | `/v1/runs/{id}/events` | Event list |
| GET | `/v1/runs/{id}/events/stream` | SSE (dashboard proxies as `/runs/{id}/live`) |
| GET | `/v1/runs/compare?base=&head=` | Diff of findings |

Health: `GET /healthz`.

---

## Schedules & notify (F4)

Optional cron enqueue:

```bash
curl -sS -X POST "$API/v1/schedules" -H 'Content-Type: application/json' \
  -d '{"cron":"0 * * * *","preset":"quality","enabled":true}'
```

Notify via Slack webhook and/or Telegram when a run finishes, filtered by `LAB_NOTIFY_ON` (e.g. `fail`, `warn+fail`).

Cron expression primer: [crontab.guru](https://crontab.guru/).

---

## Worker model (101)

- Default **1 worker** — jobs run **serially** from an in-process queue.  
- Prefer async enqueue for long scenarios; poll or watch the dashboard.  
- Job context timeout is on the order of **60 minutes** (see isolation doc).  
- Memory store loses history on API restart; set `LAB_DATABASE_URL` / `DATABASE_URL` for Postgres.

Isolation policy: [.project/vps/f2-isolation.md](../.project/vps/f2-isolation.md).

---

## Compare runs

Dashboard: `/compare`  
API: `/v1/runs/compare?base=<id>&head=<id>`

Use after fixing a theme zip — same preset, new zip binding — to see added/removed findings.

Next: [06 — Configuration](./06-configuration.md).
