# Anti-patterns

Reject PRs that reintroduce:

1. **Checks inside domain** — no Docker, no WPScan, no Lighthouse imports in `packages/domain`
2. **WordPress logic in core** — WP belongs in `adapters/wordpress` and org/sec runners
3. **Rewriting tools in Go** — wrap Lighthouse/vnu/axe/WPScan as runners
4. **Plugin territory in theme zips** — login limiters, WAF, hide-login for .org themes stay in site baseline, not theme pack
5. **Coupling Lab to wpfasty** — this repo must run without the theme monorepo
6. **One-off scripts without Manifest** — every lab run is a Manifest → Report
7. **Security scans of third-party sites** without explicit opt-in and legal clarity
8. **Duplicating policy in runners** — runners emit findings; policy maps baskets
9. **Hardcoding only org/sec/quality** — framework must allow new lab ids without core forks
10. **Admin-bar / debug-on quality runs** — quality lab must use prod-like, logged-out targets
