# 01 — Introduction

## What is FastyGo Lab?

FastyGo Lab is an **automated test laboratory** for WordPress themes (and related static sites). You give it:

1. a **theme zip** (what you would upload to wordpress.org or ship to a customer), and/or  
2. a **running site URL** (usually a disposable WordPress in Docker),

and it returns a machine-readable **Report**: findings, severities, and **decision baskets** that answer *who should fix this — the theme, the site baseline, or is it acceptable?*

It is **not** a substitute for human theme review, security consultants, or Lighthouse “green scores as marketing.” It is a **repeatable lab** so juniors and seniors share the same gates.

---

## The three product questions

| Scenario | Product question | Lab preset |
|----------|------------------|------------|
| **Org** | Will wordpress.org Theme Review reject this zip? | `org` |
| **Quality** | Is it fast, valid, accessible, and crawlable enough? | `quality` / `quality-wp` |
| **Security** | What should we cut from the theme vs harden on the site? | `sec` |

Design intent for each lives in:

- [.project/check/theme-check.md](../.project/check/theme-check.md)  
- [.project/check/validate-check.md](../.project/check/validate-check.md)  
- [.project/check/security-check.md](../.project/check/security-check.md)  

Implementation status: [.project/check/audit-progress.md](../.project/check/audit-progress.md).

---

## Two ways to run the same work

| Mode | Entry | Best for |
|------|-------|----------|
| **Local CLI** | `go run ./apps/cli run -f …` or `make org` | Debugging runners, adding checks, CI on a laptop |
| **SaaS API + dashboard** | `POST /v1/runs` + browser UI | Shared lab host, demos, schedules, watching live timelines |

Both use the same **Manifest → Orchestrator → Report** contract. SaaS adds a job queue, event stream, and UI.

```text
Manifest  →  Adapter (serve / attach site)
          →  Gates → Checks → Runners (often Docker)
          →  Policy (Finding → Decision)
          →  Report (+ events for the dashboard)
```

---

## Who should read what

| Role | Focus |
|------|--------|
| **Intern / junior** | This intro + [concepts](./02-concepts.md) + [quickstart](./03-quickstart-local.md) + run `demo` |
| **Theme / frontend engineer** | [Scenarios](./04-scenarios.md) — especially org + quality |
| **Backend / platform** | [API](./05-api-dashboard.md) + [configuration](./06-configuration.md) |
| **Security-minded** | Scenario `sec` + OWASP links in [further reading](./09-further-reading.md) |
| **Team lead** | Decision baskets, golden exits (`pass`/`fail`/`warn`/`error`), isolation notes |

---

## Mental model for beginners

Think of Lab like a **CI pipeline specialized for themes**:

1. **Zip lint** — packaging rules (cheap, no WordPress).  
2. **Theme Check** — official-style WordPress theme rules in a clean install.  
3. **HTTP matrix** — hit front/single/404/… and fail on PHP notices.  
4. **Lighthouse / axe / vnu** — quality of the rendered pages.  
5. **WPScan / nuclei / static PHP** — security posture of theme + lab site.

Each step emits **findings**. Policy turns findings into **ACCEPT / FIX_THEME / SITE_DEFAULT_ON / …** so the report is actionable, not just a wall of tool output.

---

## What Lab deliberately does *not* do

- Guarantee wordpress.org acceptance (human review still exists).  
- Grant the `accessibility-ready` tag without deeper a11y work.  
- Attack arbitrary third-party sites (sec targets must be **owned / allowlisted**).  
- Bundle WAF / login-limiters into a .org theme zip (plugin / site territory).

Next: [02 — Core concepts](./02-concepts.md).
