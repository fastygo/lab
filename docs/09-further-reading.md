# 09 — Further reading

Curated links for the stacks Lab wraps. Prefer primary docs over random blog posts.

---

## WordPress & themes

| Resource | Use |
|----------|-----|
| [Theme Handbook](https://make.wordpress.org/themes/handbook/) | Org review culture |
| [Theme Handbook — Required](https://make.wordpress.org/themes/handbook/review/required/) | Packaging & forbidden practices |
| [Theme Check](https://wordpress.org/plugins/theme-check/) | Plugin Lab automates headlessly |
| [Theme Unit Test / test data](https://github.com/WordPress/theme-test-data) | Content matrix |
| [WP-CLI](https://wp-cli.org/) | How seed/install scripts work |
| [Developing with WordPress](https://developer.wordpress.org/) | Hooks, escaping, APIs |

---

## Quality / performance / a11y

| Resource | Use |
|----------|-----|
| [Lighthouse documentation](https://developer.chrome.com/docs/lighthouse/overview/) | Q1 scores & audits |
| [web.dev — Core Web Vitals](https://web.dev/articles/vitals) | LCP, INP, CLS |
| [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci) | Assert budgets in CI |
| [WCAG 2.2](https://www.w3.org/TR/WCAG22/) | Accessibility bar |
| [axe-core](https://github.com/dequelabs/axe-core) | Automated a11y rules Lab uses |
| [Nu Html Checker](https://validator.w3.org/nu/) | HTML5 validation |
| [Playwright](https://playwright.dev/docs/intro) | Keyboard / browser gates |

---

## Security

| Resource | Use |
|----------|-----|
| [OWASP Top 10](https://owasp.org/www-project-top-ten/) | Vocabulary for seniors & juniors |
| [OWASP ASVS](https://owasp.org/www-project-application-security-verification-standard/) | Deeper verification levels |
| [WordPress Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/WordPress_Security_Cheat_Sheet.html) | Site baseline ideas |
| [WPScan](https://github.com/wpscanteam/wpscan) | Enum / CVE scanning |
| [Nuclei](https://docs.projectdiscovery.io/tools/nuclei/overview) | Template-driven checks |
| [PHP Security Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/PHP_Configuration_Cheat_Sheet.html) | Runtime hardening |
| [CWE Top 25](https://cwe.mitre.org/top25/) | Finding taxonomy |

---

## Go & platform

| Resource | Use |
|----------|-----|
| [Go modules reference](https://go.dev/ref/mod) | This monorepo’s module |
| [templ](https://templ.guide/) | Dashboard templates (`apps/web`) |
| [Docker Compose](https://docs.docker.com/compose/) | Lab profiles |
| [Server-Sent Events (MDN)](https://developer.mozilla.org/en-US/docs/Web/API/Server-sent_events) | Live run timeline |
| [PostgreSQL](https://www.postgresql.org/docs/current/) | Optional run store |

---

## Hexagonal / DDD (seniors mentoring juniors)

| Resource | Use |
|----------|-----|
| [Hexagonal Architecture (Alistair Cockburn)](https://alistair.cockburn.us/hexagonal-architecture/) | Ports & adapters mindset of Lab |
| [Domain-Driven Design Reference (Vernon)](https://www.domainlanguage.com/ddd/reference/) | Bounded contexts vocabulary |

Lab’s own map: [.project/architecture.md](../.project/architecture.md).

---

## Internal checklists (design → code)

| Doc | Maps to |
|-----|---------|
| [.project/check/theme-check.md](../.project/check/theme-check.md) | Org scenario design |
| [.project/check/validate-check.md](../.project/check/validate-check.md) | Quality scenario design |
| [.project/check/security-check.md](../.project/check/security-check.md) | Sec scenario design |
| [.project/check/audit-progress.md](../.project/check/audit-progress.md) | Implemented coverage |
| [.project/vps/cycle-f-saas.md](../.project/vps/cycle-f-saas.md) | API / dashboard / notify |

Back to [README](./README.md) · [Glossary](./glossary.md).
