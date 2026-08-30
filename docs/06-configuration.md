# 06 — Configuration

Environment variables are the control plane. Example file: [`deploy/compose/.env.example`](../deploy/compose/.env.example). On the shared VPS, SaaS services often load `/etc/fastygo-lab-saas.env`.

---

## WordPress / compose

| Variable | Purpose |
|----------|---------|
| `LAB_WP_URL` | Public site URL written into WP (`WP_HOME` / `WP_SITEURL`). Wrong value → assets and redirects break in browsers. |
| Compose profiles | `org`, `quality`, `saas` via `docker compose --profile …` |

See also [WordPress `WP_HOME` / `WP_SITEURL`](https://developer.wordpress.org/advanced-administration/wordpress/wp-config/#wp-siteurl).

---

## API / worker

| Variable | Default / notes |
|----------|-----------------|
| `LAB_REPO_ROOT` | Repo root on the worker (Manifest + zip paths resolve here) |
| `LAB_API_ADDR` | Bind address (e.g. `:8090`) |
| `LAB_API_WORKERS` | Concurrent workers (default `1`) |
| `LAB_DATABASE_URL` or `DATABASE_URL` | Postgres URI; else memory store |
| `LAB_ALLOWED_BASE_URLS` | Comma-separated URL prefixes for wordpress adapter (sec safety) |
| `LAB_WP_PASSWORD` | Optional — auth-abuse cookie checks |
| `WPSCAN_API_TOKEN` | Optional — richer WPScan CVE results ([WPScan API](https://wpscan.com/api)) |
| `LAB_NUCLEI_TAGS` | Optional nuclei tag filter |

---

## Dashboard (`lab-web`)

| Variable | Purpose |
|----------|---------|
| `LAB_WEB_ADDR` | Bind (local `:8091`, some hosts `:8092`) |
| `LAB_API_URL` | **Internal** API base for SSR + SSE proxy (loopback OK) |
| `LAB_API_PUBLIC_URL` | **Browser-facing** API origin for links; if empty, derived from request Host + API port |
| `LAB_DASHBOARD_URL` | Used in notify messages (human link back to UI) |

---

## Notify / scheduler

| Variable | Purpose |
|----------|---------|
| `SLACK_WEBHOOK_URL` | Incoming webhook |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | Bot notify |
| `LAB_NOTIFY_ON` | e.g. `fail`, `warn+fail` |
| `LAB_SCHEDULER_INTERVAL` | Tick interval for cron schedules |

Slack incoming webhooks: [api.slack.com/messaging/webhooks](https://api.slack.com/messaging/webhooks).  
Telegram bots: [core.telegram.org/bots](https://core.telegram.org/bots).

---

## Preset bindings (API body)

Applied in `packages/presets` onto the loaded Manifest:

| JSON field | Effect |
|------------|--------|
| `themeZip` | Adapter + zip-aware checks |
| `baseUrl` | Adapter target URL |
| `root` | Static adapters |
| `config` | Adapter config merge |
| `checkConfig` | Check config merge (e.g. `dockerNetwork`) |

Paths are snapshotted relative to `LAB_REPO_ROOT`.

---

## Policy packs

Manifest `spec.policy.pack` selects classification rules, e.g.:

- `wordpress-org` — org lab  
- `lightspeed` — quality  
- `secure-baseline` — sec  
- `default` — generic  

Seniors extending baskets should start in `packages/policy` and keep codes stable for compare/regression.

---

## Runner Docker networking (cheat sheet)

| Pattern | When |
|---------|------|
| Compose network + WP volume | Theme Check / notice-hunter talking to `http://wordpress` |
| `dockerNetwork: host` + `127.0.0.1:8080` | Lighthouse, axe, Playwright, WPScan — need host-reachable URL after WP redirects |

Wrong network = flaky redirects and false fails. Details: [.project/vps/README.md](../.project/vps/README.md).

Next: [07 — CLI cheat sheet](./07-cli.md).
