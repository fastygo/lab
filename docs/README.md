# FastyGo Lab — Onboarding Guide (101)

Welcome. This folder is the **product textbook** for FastyGo Lab: what it is, how to run it, and how the three product scenarios map to real WordPress / web quality work.

It is written for a mixed team — interns through seniors. Start at the top; skip chapters you already know.

---

## Learning path

| Step | Who | Document | Time |
|------|-----|----------|------|
| 1 | Everyone | [01 — Introduction](./01-introduction.md) | 10 min |
| 2 | Everyone | [02 — Core concepts](./02-concepts.md) | 15 min |
| 3 | Engineers setting up a laptop | [03 — Local quickstart](./03-quickstart-local.md) | 30–60 min |
| 4 | Everyone shipping themes | [04 — Product scenarios](./04-scenarios.md) | 20 min |
| 5 | API / dashboard users | [05 — API & dashboard](./05-api-dashboard.md) | 20 min |
| 6 | Ops / seniors | [06 — Configuration](./06-configuration.md) | 15 min |
| 7 | Day-to-day CLI | [07 — CLI cheat sheet](./07-cli.md) | 10 min |
| 8 | Host / VPS operators | [08 — Operations](./08-operations.md) | 20 min |
| 9 | Deep dives | [09 — Further reading](./09-further-reading.md) | as needed |
| — | Reference | [Glossary](./glossary.md) | — |

---

## One-sentence summary

**Lab** runs automated **gates** (zip lint, Theme Check, Lighthouse, WPScan, …) against a **theme zip** or **site URL**, then returns a structured **Report** with findings and decision baskets (fix the theme vs harden the site).

Same engine locally (`lab` CLI) and on the SaaS API / dashboard.

---

## Live lab (current VPS)

| Surface | URL |
|---------|-----|
| Dashboard | http://5.129.242.217:8092/ |
| API | http://5.129.242.217:8090/ |
| API health | http://5.129.242.217:8090/healthz |
| WordPress under test | http://5.129.242.217:8080/ |

> IP and ports are environment-specific. Prefer env vars `LAB_DASHBOARD_URL` / `LAB_API_PUBLIC_URL` over hardcoding.

---

## Internal design docs (advanced)

These live under `.project/` and are for roadmap / audit depth, not day-1 onboarding:

| Doc | Topic |
|-----|--------|
| [`.project/architecture.md`](../.project/architecture.md) | Hexagonal layout, packages |
| [`.project/check/theme-check.md`](../.project/check/theme-check.md) | Org / wordpress.org checklist design |
| [`.project/check/validate-check.md`](../.project/check/validate-check.md) | Quality / Lighthouse design |
| [`.project/check/security-check.md`](../.project/check/security-check.md) | Security lab design |
| [`.project/check/audit-progress.md`](../.project/check/audit-progress.md) | What is implemented vs planned |
| [`.project/vps/README.md`](../.project/vps/README.md) | VPS runbook |
| [`.project/vps/cycle-f-saas.md`](../.project/vps/cycle-f-saas.md) | SaaS roadmap (Cycle F) |

---

## How to contribute to these docs

- Keep examples copy-pasteable.
- Prefer **why** before **flags**.
- Link external standards (W3C, Theme Handbook, OWASP) instead of re-copying them.
- Update [audit-progress](../.project/check/audit-progress.md) when a check lands in code — not only these pages.
