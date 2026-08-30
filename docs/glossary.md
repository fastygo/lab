# Glossary

| Term | Short definition |
|------|------------------|
| **Adapter** | Component that prepares a target site or fixture for checks |
| **Basket / decision** | Policy classification of a finding (fix theme vs site vs accept) |
| **Check** | One named verification inside a gate |
| **Event** | Lifecycle signal (`check.started`, …) for timelines / SSE |
| **Finding** | One structured result from a runner |
| **Gate** | Ordered stage of related checks |
| **Lab** | Product scenario id (`org`, `quality`, `sec`) or the whole system |
| **Manifest** | YAML describing adapter + gates + policy |
| **Orchestrator** | Application service that runs a Manifest → Report |
| **Policy pack** | Rules that turn findings into decisions / job status |
| **Preset** | API-friendly name for a Manifest path |
| **Report** | Final JSON (and optional Markdown) for a run |
| **Runner** | Executable check implementation (often Docker) |
| **SSE** | Server-Sent Events — live event stream to the browser |
| **Theme zip** | Packaged theme root ready for install / upload |
| **Worker** | API-side process that dequeues run ids and executes orchestrator |

---

## Status words (do not confuse)

| Word | Means |
|------|--------|
| **pass** | Audit completed; policy happy |
| **fail** | Audit completed; policy unhappy |
| **warn** | Audit completed; soft issues |
| **error** | Lab broke before a trustworthy Report |
| **queued / running** | Job lifecycle on the API |

---

## Acronyms

| Acronym | Expansion |
|---------|-----------|
| **CWV** | Core Web Vitals |
| **LH** | Lighthouse |
| **SSR** | Server-side rendering (dashboard HTML) |
| **VNU** | Nu Html Checker (HTML validator) |
| **WCAG** | Web Content Accessibility Guidelines |
| **WP** | WordPress |
