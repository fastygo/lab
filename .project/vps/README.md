# VPS deployment — clean host from source

English runbook for bringing **FastyGo Lab** up on a fresh Debian/Ubuntu VPS with **no interactive server babysitting** after the first bootstrap.  
Path convention on the lab host: `/opt/fastygo/lab`.

---

## Honest status (can you deploy “from sources only”?)

| Layer | Automated today? | Notes |
|-------|------------------|-------|
| OS packages (Docker, Compose, Git, Make, unzip, curl) | **Partial** | Not in-repo as one idempotent installer yet — see §1 |
| Go toolchain | **Partial** | Install Go 1.22+ (project uses 1.26.x) once per host |
| Clone / rsync repo | **Manual once** | SSH + git clone or `scp`/`rsync` |
| Compose org + quality stacks | **Yes** | `deploy/bootstrap-server.sh` + `make org-up` / `quality-up` |
| Runner images | **Yes** | `make runners` (long first build; needs Docker Hub access) |
| WP public URL (`LAB_WP_URL`) | **Semi** | Must set to public IP/DNS or assets point at `127.0.0.1` |
| Theme Unit Test seed + primary menu | **Semi** | `make org-seed` + menu assign (seed script now assigns `primary`) |
| Firewall / SSH keys / DNS | **Always manual** | Hosting-specific |
| Docker Hub rate limits | **Ops** | Mirrors or authenticated pulls may be required |

**Bottom line:** After OS + Go + Docker are installed and the repo is on disk, **lab stack + runners + first `lab run` are scriptable**. A true one-command “empty VPS → green org+quality” installer is **not** complete yet; this doc is the missing playbook.

---

## What required heavy manual work in practice

During the first lab-server bring-up, most pain was **host bootstrap and Docker reality**, not Lab Go code:

1. **Docker Engine + Compose plugin** on Debian, daemon restart, permission for the deploy user.
2. **Go install** (tar under `/usr/local/go`, `PATH`).
3. **Docker Hub 429** — registry mirrors / retry; runner builds pull Playwright, Node, WordPress, MySQL, Chromium.
4. **`LAB_WP_URL`** — without public `http://<VPS-IP>:8080`, theme CSS/JS URLs break for remote browsers.
5. **CRLF on Windows → Linux** — `sed` strip on `runners/**/*.sh` (bootstrap does this).
6. **Theme zip path separators** — Windows-packed zips with `\` broke unzip; use POSIX paths in `testdata/dist/*.zip`.
7. **Latte cache ownership** — Apache uid **33**, not CLI `www-data` 82.
8. **Primary menu empty** after Unit Test import until assigned to location `primary` (now in seed).
9. **Playwright vs `http://wordpress`** — WP redirects to public URL; browser runners use `dockerNetwork: host` + `http://127.0.0.1:8080`.
10. **vnu JSON on stderr** — fixed in runner; rebuild `lab/vnu:local` after pull.

Once those are known, a second VPS should follow this file with little improvisation.

---

## Prerequisites (VPS)

- Debian 12 / Ubuntu 22.04+ (x86_64)
- Root or sudo
- Open ports: **22** (SSH), **8080** (WP), **8091** (quality fixture, optional)
- ≥ 2 vCPU, ≥ 4 GB RAM (8 GB comfortable for Lighthouse ×3 + Playwright)
- ≥ 40 GB disk (runner images + Chromium + WP volume)

---

## 1. Host packages (one-time)

```bash
apt-get update
apt-get install -y ca-certificates curl git make unzip jq

# Docker Engine + Compose plugin (official convenience script or distro docs)
curl -fsSL https://get.docker.com | sh
systemctl enable --now docker
docker compose version

# Go (match go.mod; example 1.26.4)
curl -fsSL https://go.dev/dl/go1.26.4.linux-amd64.tar.gz -o /tmp/go.tgz
rm -rf /usr/local/go && tar -C /usr/local -xzf /tmp/go.tgz
echo 'export PATH=/usr/local/go/bin:$PATH' >/etc/profile.d/golang.sh
source /etc/profile.d/golang.sh
go version
```

Optional: configure Docker Hub mirror under `/etc/docker/daemon.json` if you hit rate limits, then `systemctl restart docker`.

---

## 2. Place the Lab source tree

```bash
mkdir -p /opt/fastygo
# Option A — git
git clone <YOUR_LAB_REMOTE> /opt/fastygo/lab
# Option B — from a workstation
# rsync -az --delete ./lab/ root@VPS:/opt/fastygo/lab/

cd /opt/fastygo/lab
test -f go.mod && test -f deploy/bootstrap-server.sh
```

Ensure `testdata/dist/latte.zip` exists (POSIX paths inside the zip).

---

## 3. Public WordPress URL

```bash
cd /opt/fastygo/lab
cp deploy/compose/.env.example deploy/compose/.env
# Edit LAB_WP_URL to the address browsers and Playwright can reach, e.g.:
# LAB_WP_URL=http://YOUR.VPS.IP:8080
```

Export for the current shell too:

```bash
export LAB_WP_URL=http://YOUR.VPS.IP:8080
export PATH=/usr/local/go/bin:$PATH
```

---

## 4. Bootstrap Compose + theme

```bash
cd /opt/fastygo/lab
bash deploy/bootstrap-server.sh
```

This starts org (MySQL + WordPress + wpcli) and quality nginx fixture, activates latte when the zip is present, and sets `siteurl`/`home` from `LAB_WP_URL`.

Verify:

```bash
curl -sI http://127.0.0.1:8080/ | head -5
curl -sI "$LAB_WP_URL/" | head -5
docker compose -f deploy/compose/docker-compose.yml --profile org ps
```

---

## 5. Build runner images

```bash
cd /opt/fastygo/lab
make runners
# First build can take 15–40+ minutes depending on bandwidth.
docker images | grep '^lab/'
```

Expected images include: `lab/lighthouse`, `lab/axe`, `lab/theme-check`, `lab/vnu`, `lab/notice-hunter`, `lab/org-keyboard`, `lab/css-lint`, `lab/quality-extras`, `lab/wpscan`, `lab/composer-audit`, `lab/nuclei`, `lab/phpcs-security`, `lab/semgrep`.

---

## 6. Seed Theme Unit Test data (org)

```bash
cd /opt/fastygo/lab
make org-seed
# Confirms fixtures XML mount + org-seed.json on host
cat testdata/fixtures/org-seed.json
```

Seed assigns a primary menu when menus exist (needed for nav/sheet keyboard tests).

---

## 7. Smoke each lab

```bash
cd /opt/fastygo/lab
export PATH=/usr/local/go/bin:$PATH

go run ./apps/cli labs
go run ./apps/cli run -f testdata/manifests/demo.lab.yaml

# Org (zip + Theme Check + matrix + notices + keyboard)
go run ./apps/cli run -f testdata/manifests/org.lab.yaml -o /tmp/org.audit.json

# Quality on static fixture
go run ./apps/cli run -f testdata/manifests/quality.lab.yaml -o /tmp/quality.audit.json

# Quality against live WP (logged-out)
go run ./apps/cli run -f testdata/manifests/quality-wp.lab.yaml -o /tmp/quality-wp.audit.json

# Sec (needs WP up)
go run ./apps/cli run -f testdata/manifests/sec.lab.yaml -o /tmp/sec.audit.json
```

`status: fail` with theme findings (e.g. Theme Check required, axe) can still mean **Lab is healthy** — inspect findings, not only exit code.

---

## 8. Day-2 operations

| Task | Command |
|------|---------|
| Restart stacks | `make org-up` / `make quality-up` |
| Rebuild one runner | `docker build -t lab/vnu:local runners/vnu` |
| Re-seed | `make org-seed` |
| Update code | `git pull` then `make runners` if Dockerfiles/entrypoints changed |
| Logs | `docker compose -f deploy/compose/docker-compose.yml --profile org logs -f wordpress` |

From a workstation with wpfasty:

```bash
export LAB_ROOT=/opt/fastygo/lab   # or path via SSH mount / sync
bun run theme:audit -- latte
# writes themes/latte/dist/latte.audit.json via lab org run -o
```

---

## 9. Gaps to close for “zero-touch” (future)

Track these if you want a single `curl | bash` or cloud-init module:

1. `deploy/install-host.sh` — apt + Docker + Go (idempotent).
2. `deploy/cloud-init.yaml` — clone repo, write `.env`, run bootstrap + `make runners`.
3. Documented Hub auth / mirror defaults for CI VPS images.
4. Healthcheck target: `lab run` exit 0 on a golden theme fixture (not latte-with-known-fails).
5. Non-root deploy user + rootless Docker notes.

Until then, treat **§1–§7** as the supported clean-VPS path.

---

## Cycle F — SaaS on this host

Product roadmap (API, workers, dashboard events, schedules, Telegram/Slack):

→ **[cycle-f-saas.md](./cycle-f-saas.md)**

Workers reuse images from `make runners` and compose profiles from this runbook.

---

## Quick checklist (print / paste)

- [ ] Debian/Ubuntu VPS, ports 22/8080/(8091)
- [ ] Docker + `docker compose` + Make + Git + unzip
- [ ] Go on `PATH`
- [ ] Repo at `/opt/fastygo/lab` with `testdata/dist/latte.zip`
- [ ] `deploy/compose/.env` → `LAB_WP_URL=http://<public>:8080`
- [ ] `bash deploy/bootstrap-server.sh`
- [ ] `make runners`
- [ ] `make org-seed`
- [ ] `go run ./apps/cli run -f testdata/manifests/org.lab.yaml -o /tmp/org.audit.json`
- [ ] `go run ./apps/cli run -f testdata/manifests/quality.lab.yaml -o /tmp/quality.audit.json`

Related: [../constructor/compose-profiles.md](../constructor/compose-profiles.md), [../../deploy/bootstrap-server.sh](../../deploy/bootstrap-server.sh), [../labs/quality.md](../labs/quality.md), [cycle-f-saas.md](./cycle-f-saas.md).
