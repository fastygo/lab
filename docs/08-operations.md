# 08 — Operations

For people who keep the shared lab host healthy. Pair with [.project/vps/README.md](../.project/vps/README.md).

---

## Day-2 checklist

1. **Disk** — runner images + Chromium + WP volumes need tens of GB.  
2. **Compose org up** — WordPress `:8080` returns 200.  
3. **`lab-api` / `lab-web` active** — systemd or process manager.  
4. **Env file** — `LAB_WP_URL`, `LAB_ALLOWED_BASE_URLS`, public API/dashboard URLs.  
5. **Zip present** — `testdata/dist/*.zip` on the worker.  
6. **Serial jobs** — do not flood mutating presets in parallel on one WP.

---

## Typical systemd layout (example)

```text
lab-api.service   → /usr/local/bin/lab-api   EnvironmentFile=/etc/fastygo-lab-saas.env
lab-web.service   → /usr/local/bin/lab-web   same env file
```

Restart after binary deploy:

```bash
systemctl restart lab-api lab-web
curl -sf http://127.0.0.1:8090/healthz
curl -sf http://127.0.0.1:8092/healthz
```

---

## Persistence

| Store | Behavior |
|-------|----------|
| Memory (default) | Fast; **history wiped** on API restart |
| Postgres | Set `LAB_DATABASE_URL` / `DATABASE_URL`; runs/events survive restarts |

Compose profile `saas` starts local Postgres for development.

---

## Security ops

- Keep `LAB_ALLOWED_BASE_URLS` set on any internet-reachable API.  
- Do not expose WP admin with default `wp/wp` outside a lab VLAN.  
- Treat `WPSCAN_API_TOKEN`, Slack, Telegram tokens as secrets (env only).  
- Sec jobs may **fail** loudly — that is the product; alert on `error` for infra.

OWASP hosting guidance: [WordPress Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/WordPress_Security_Cheat_Sheet.html).

---

## Observability 101

| Signal | Where |
|--------|-------|
| API health | `/healthz` (`store`, `cycle`, version) |
| Web health | `/healthz` (`api`, `apiPublic`) |
| Job progress | Dashboard timeline / SSE |
| Worker logs | `journalctl -u lab-api` (or stderr of `go run`) |

---

## Failure playbook (short)

| Symptom | Check |
|---------|--------|
| Runs stuck `queued` | Is worker goroutine alive? Queue full? API crash loop? |
| All browsers fail | `LAB_WP_URL` / redirects / port 8080 |
| Theme Check empty/fail infra | `make runners`, compose network/volume names |
| Dashboard shows `127.0.0.1` links | Set `LAB_API_PUBLIC_URL` to public API origin |
| History empty after reboot | Enable Postgres URL |

Next: [09 — Further reading](./09-further-reading.md).
