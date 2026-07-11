package policy

import "github.com/fastygo/lab/packages/domain"

// Rule maps a finding code to a decision basket.
type Rule struct {
	Code      string
	Basket    domain.Basket
	Rationale string
}

// Engine maps findings to decisions using a pack of rules.
type Engine struct {
	pack  string
	rules map[string]Rule
}

// NewEngine builds an engine for a named pack.
func NewEngine(pack string) *Engine {
	if pack == "" {
		pack = "default"
	}
	e := &Engine{pack: pack, rules: map[string]Rule{}}
	for _, r := range packRules(pack) {
		e.rules[r.Code] = r
	}
	return e
}

func packRules(pack string) []Rule {
	switch pack {
	case "lightspeed":
		return append(defaultRules(), []Rule{
			{Code: "quality.lighthouse.performance", Basket: domain.BasketBudget, Rationale: "Tune perf budgets / FIX_THEME assets"},
			{Code: "quality.lighthouse.accessibility", Basket: domain.BasketFixTheme, Rationale: "A11y score below threshold"},
			{Code: "quality.axe.ok", Basket: domain.BasketAccept, Rationale: "No axe violations"},
			{Code: "quality.vnu.ok", Basket: domain.BasketAccept, Rationale: "HTML validates"},
			{Code: "quality.vnu.no_errors", Basket: domain.BasketAccept, Rationale: "No vnu errors"},
			{Code: "quality.css.ok", Basket: domain.BasketAccept, Rationale: "CSS parse clean"},
			{Code: "quality.css.summary", Basket: domain.BasketAccept, Rationale: "CSS lint summary"},
			{Code: "quality.seo.ok", Basket: domain.BasketAccept, Rationale: "SEO meta OK"},
			{Code: "quality.seo.summary", Basket: domain.BasketAccept, Rationale: "SEO meta summary"},
			{Code: "quality.seo.social_skipped", Basket: domain.BasketAccept, Rationale: "Social graph optional"},
			{Code: "quality.seo.title_ok", Basket: domain.BasketAccept, Rationale: "Title present"},
			{Code: "quality.seo.viewport_ok", Basket: domain.BasketAccept, Rationale: "Viewport present"},
			{Code: "quality.seo.h1_ok", Basket: domain.BasketAccept, Rationale: "Single h1"},
			{Code: "quality.seo.description_ok", Basket: domain.BasketAccept, Rationale: "Description present"},
			{Code: "quality.seo.description_missing", Basket: domain.BasketAccept, Rationale: "Description soft"},
			{Code: "quality.extras.ok", Basket: domain.BasketAccept, Rationale: "Viewports + console clean"},
			{Code: "quality.extras.summary", Basket: domain.BasketAccept, Rationale: "Quality extras summary"},
			{Code: "quality.viewport.ok", Basket: domain.BasketAccept, Rationale: "Viewport rendered"},
			{Code: "quality.console.ok", Basket: domain.BasketAccept, Rationale: "Console clean"},
			{Code: "quality.motion.ok", Basket: domain.BasketAccept, Rationale: "Reduced motion OK"},
			{Code: "quality.links.ok", Basket: domain.BasketAccept, Rationale: "Broken-link crawl clean"},
			{Code: "quality.links.summary", Basket: domain.BasketAccept, Rationale: "Broken-link summary"},
			{Code: "quality.links.none", Basket: domain.BasketAccept, Rationale: "No same-origin links"},
			{Code: "quality.lighthouse.summary", Basket: domain.BasketAccept, Rationale: "Lighthouse median summary"},
			{Code: "quality.vnu.soft_error", Basket: domain.BasketBudget, Rationale: "Unit Test / content HTML soft mode"},
			{Code: "quality.seo.og_ok", Basket: domain.BasketAccept, Rationale: "og:title present"},
			{Code: "quality.seo.og_type_ok", Basket: domain.BasketAccept, Rationale: "og:type present"},
			{Code: "quality.seo.og_url_ok", Basket: domain.BasketAccept, Rationale: "og:url present"},
			{Code: "quality.seo.twitter_ok", Basket: domain.BasketAccept, Rationale: "twitter:card present"},
			{Code: "quality.seo.jsonld_ok", Basket: domain.BasketAccept, Rationale: "JSON-LD valid"},
			{Code: "runner.docker.unavailable", Basket: domain.BasketAccept, Rationale: "Docker missing in unit/dev; re-run with Docker for real scores"},
		}...)
	case "wordpress-org":
		return append(defaultRules(), []Rule{
			{Code: "org.zip.ok", Basket: domain.BasketAccept, Rationale: "Packaging OK"},
			{Code: "org.zip.tag_accessibility_ready", Basket: domain.BasketBlockTag, Rationale: "Do not claim accessibility-ready yet"},
			{Code: "org.zip.missing_style_css", Basket: domain.BasketFixTheme, Rationale: "Required theme file"},
			{Code: "org.zip.missing_readme_txt", Basket: domain.BasketFixTheme, Rationale: "Required for .org"},
			{Code: "org.zip.missing_license", Basket: domain.BasketFixTheme, Rationale: "GPL license file required"},
			{Code: "org.matrix.listed", Basket: domain.BasketAccept, Rationale: "Matrix recorded"},
			{Code: "org.matrix.ok", Basket: domain.BasketAccept, Rationale: "HTTP smoke OK"},
			{Code: "org.matrix.smoke_summary", Basket: domain.BasketAccept, Rationale: "HTTP smoke summary"},
			{Code: "org.notice.ok", Basket: domain.BasketAccept, Rationale: "No theme debug notices"},
			{Code: "org.notice.summary", Basket: domain.BasketAccept, Rationale: "Notice hunter summary"},
			{Code: "org.keyboard.ok", Basket: domain.BasketAccept, Rationale: "Keyboard scenarios passed"},
			{Code: "org.keyboard.summary", Basket: domain.BasketAccept, Rationale: "Keyboard smoke summary"},
			{Code: "org.keyboard.skip_ok", Basket: domain.BasketAccept, Rationale: "Skip link OK"},
			{Code: "org.keyboard.nav_ok", Basket: domain.BasketAccept, Rationale: "Primary nav keyboard OK"},
			{Code: "org.keyboard.sheet_ok", Basket: domain.BasketAccept, Rationale: "Mobile sheet keyboard OK"},
			{Code: "org.keyboard.search_ok", Basket: domain.BasketAccept, Rationale: "Search keyboard OK"},
			{Code: "org.themecheck.ok", Basket: domain.BasketAccept, Rationale: "Theme Check clean"},
			{Code: "org.themecheck.no_required", Basket: domain.BasketAccept, Rationale: "No Theme Check required errors"},
			{Code: "org.themecheck.plugin_ready", Basket: domain.BasketAccept, Rationale: "Theme Check installed"},
			{Code: "runner.docker.unavailable", Basket: domain.BasketAccept, Rationale: "Theme Check needs Docker compose org profile"},
		}...)
	case "secure-baseline":
		return append(defaultRules(), []Rule{
			{Code: "sec.headers.ok", Basket: domain.BasketAccept, Rationale: "Headers OK"},
			{Code: "sec.headers.nosniff", Basket: domain.BasketSiteDefaultOn, Rationale: "Enable nosniff on site baseline"},
			{Code: "sec.headers.clickjacking", Basket: domain.BasketSiteDefaultOn, Rationale: "Enable frame protections"},
			{Code: "sec.headers.referrer", Basket: domain.BasketSiteDefaultOn, Rationale: "Set Referrer-Policy"},
			{Code: "sec.headers.csp", Basket: domain.BasketSiteDefaultOn, Rationale: "Start CSP report-only on site baseline"},
			{Code: "sec.headers.hsts", Basket: domain.BasketSiteDefaultOn, Rationale: "Enable HSTS on HTTPS prod only"},
			{Code: "sec.headers.permissions", Basket: domain.BasketSiteDefaultOn, Rationale: "Set Permissions-Policy on site baseline"},
			{Code: "sec.config.file_edit", Basket: domain.BasketSiteDefaultOn, Rationale: "Set DISALLOW_FILE_EDIT on site baseline"},
			{Code: "sec.recon.xmlrpc", Basket: domain.BasketSiteDefaultOff, Rationale: "Disable xmlrpc on site baseline"},
			{Code: "sec.recon.readme", Basket: domain.BasketSiteDefaultOff, Rationale: "Remove readme.html from deploy"},
			{Code: "sec.recon.license", Basket: domain.BasketSiteDefaultOff, Rationale: "Remove license.txt from webroot if desired"},
			{Code: "sec.recon.generator", Basket: domain.BasketSiteDefaultOff, Rationale: "Remove generator meta on site baseline"},
			{Code: "sec.recon.user_enum.author", Basket: domain.BasketSiteDefaultOff, Rationale: "Block author enumeration"},
			{Code: "sec.recon.user_enum.rest", Basket: domain.BasketSiteDefaultOff, Rationale: "Restrict REST users endpoint"},
			{Code: "sec.recon.rest_index", Basket: domain.BasketAccept, Rationale: "REST index needed for Gutenberg — document risk"},
			{Code: "sec.recon.registration", Basket: domain.BasketSiteDefaultOff, Rationale: "users_can_register=0"},
			{Code: "sec.recon.dir_listing.uploads", Basket: domain.BasketSiteDefaultOn, Rationale: "Options -Indexes on uploads"},
			{Code: "sec.recon.dir_listing.themes", Basket: domain.BasketSiteDefaultOn, Rationale: "Options -Indexes on themes"},
			{Code: "sec.recon.wp_cron", Basket: domain.BasketSiteDefaultOff, Rationale: "DISABLE_WP_CRON + system cron"},
			{Code: "sec.recon.sensitive.env", Basket: domain.BasketFixSite, Rationale: "Never deploy secrets to webroot"},
			{Code: "sec.recon.sensitive.wpconfig_bak", Basket: domain.BasketFixSite, Rationale: "Remove config backups from webroot"},
			{Code: "sec.recon.sensitive.debug_log", Basket: domain.BasketFixSite, Rationale: "Deny debug.log via web"},
			{Code: "sec.recon.sensitive.composer", Basket: domain.BasketFixSite, Rationale: "Do not expose composer.json in webroot"},
			{Code: "sec.wpscan.completed", Basket: domain.BasketAccept, Rationale: "Enumeration completed"},
			{Code: "sec.wpscan.users", Basket: domain.BasketSiteDefaultOff, Rationale: "Hide/limit user enumeration"},
			{Code: "sec.wpscan.vuln", Basket: domain.BasketFixSite, Rationale: "Patch core/plugin/theme CVE"},
			{Code: "sec.wpscan.exec_failed", Basket: domain.BasketFixSite, Rationale: "WPScan execution failed"},
			{Code: "sec.composer.ok", Basket: domain.BasketAccept, Rationale: "No Composer advisories"},
			{Code: "sec.composer.completed", Basket: domain.BasketAccept, Rationale: "Composer audit finished"},
			{Code: "sec.composer.advisory", Basket: domain.BasketFixTheme, Rationale: "Upgrade/replace vulnerable Composer package"},
			{Code: "sec.composer.abandoned", Basket: domain.BasketBudget, Rationale: "Review abandoned packages"},
			{Code: "sec.composer.lock_missing", Basket: domain.BasketAccept, Rationale: "No lockfile to audit"},
			{Code: "sec.composer.zip_missing", Basket: domain.BasketBudget, Rationale: "themeZip required for composer audit"},
			{Code: "sec.composer.exec_failed", Basket: domain.BasketFixTheme, Rationale: "composer audit failed"},
			{Code: "sec.nuclei.ok", Basket: domain.BasketAccept, Rationale: "No Nuclei wordpress matches"},
			{Code: "sec.nuclei.completed", Basket: domain.BasketAccept, Rationale: "Nuclei scan finished"},
			{Code: "sec.nuclei.match", Basket: domain.BasketFixSite, Rationale: "Address Nuclei WordPress/CVE match"},
			{Code: "runner.docker.unavailable", Basket: domain.BasketAccept, Rationale: "Docker runner image not available"},
			{Code: "sec.auth.ok", Basket: domain.BasketAccept, Rationale: "Auth abuse probes passed"},
			{Code: "sec.auth.rate_limit_present", Basket: domain.BasketAccept, Rationale: "Login rate limit observed"},
			{Code: "sec.auth.login_no_rate_limit", Basket: domain.BasketSiteDefaultOn, Rationale: "Enable login rate limit on site baseline"},
			{Code: "sec.auth.xmlrpc_multicall", Basket: domain.BasketSiteDefaultOff, Rationale: "Disable xmlrpc / block multicall brute"},
			{Code: "sec.auth.xmlrpc_rate_limited", Basket: domain.BasketAccept, Rationale: "XML-RPC rate limited"},
			{Code: "sec.auth.weak_password", Basket: domain.BasketFixSite, Rationale: "Lab account has weak password"},
			{Code: "sec.auth.cookie_no_httponly", Basket: domain.BasketSiteDefaultOn, Rationale: "Force HttpOnly on auth cookies"},
			{Code: "sec.auth.cookie_no_samesite", Basket: domain.BasketSiteDefaultOn, Rationale: "Set SameSite=Lax/Strict on auth cookies"},
			{Code: "sec.auth.cookie_samesite_none_insecure", Basket: domain.BasketSiteDefaultOn, Rationale: "SameSite=None requires Secure"},
			{Code: "sec.auth.cookie_no_secure", Basket: domain.BasketSiteDefaultOn, Rationale: "Force Secure cookies on HTTPS"},
			{Code: "sec.auth.cookie_secure_http", Basket: domain.BasketAccept, Rationale: "HTTP lab — Secure N/A until HTTPS"},
			{Code: "sec.auth.cookie_login_skipped", Basket: domain.BasketAccept, Rationale: "Optional LAB_WP_PASSWORD for cookie flags"},
			{Code: "sec.auth.login_failed", Basket: domain.BasketBudget, Rationale: "Lab credentials invalid — check LAB_WP_*"},
			{Code: "sec.auth.host_header_poison", Basket: domain.BasketFixSite, Rationale: "Harden password-reset host trust"},
			{Code: "sec.auth.host_header_ok", Basket: domain.BasketAccept, Rationale: "Host-header poison not reflected"},
			{Code: "sec.auth.login_unreachable", Basket: domain.BasketFixSite, Rationale: "wp-login unreachable"},
			{Code: "sec.theme.ok", Basket: domain.BasketAccept, Rationale: "Theme security probes passed"},
			{Code: "sec.theme.static_ok", Basket: domain.BasketAccept, Rationale: "No banned PHP patterns"},
			{Code: "sec.theme.xss_search_ok", Basket: domain.BasketAccept, Rationale: "Search XSS probe clean"},
			{Code: "sec.theme.xss_attr_ok", Basket: domain.BasketAccept, Rationale: "Attribute breakout probe clean"},
			{Code: "sec.theme.zip_missing", Basket: domain.BasketAccept, Rationale: "Static scan skipped without themeZip"},
			{Code: "sec.theme.eval", Basket: domain.BasketCutTarget, Rationale: "Remove eval from theme"},
			{Code: "sec.theme.create_function", Basket: domain.BasketCutTarget, Rationale: "Remove create_function"},
			{Code: "sec.theme.assert", Basket: domain.BasketCutTarget, Rationale: "Remove assert"},
			{Code: "sec.theme.unserialize", Basket: domain.BasketCutTarget, Rationale: "Ban unserialize on user data"},
			{Code: "sec.theme.shell", Basket: domain.BasketCutTarget, Rationale: "Remove shell execution"},
			{Code: "sec.theme.backticks", Basket: domain.BasketCutTarget, Rationale: "Remove backtick shell"},
			{Code: "sec.theme.file_get_remote", Basket: domain.BasketCutTarget, Rationale: "No remote file_get_contents in theme"},
			{Code: "sec.theme.curl_exec", Basket: domain.BasketFixTheme, Rationale: "Review curl for SSRF"},
			{Code: "sec.theme.open_redirect", Basket: domain.BasketCutTarget, Rationale: "Ban open redirect"},
			{Code: "sec.theme.echo_superglobal", Basket: domain.BasketFixTheme, Rationale: "Escape / ContextFactory only"},
			{Code: "sec.theme.sql_unprepared", Basket: domain.BasketCutTarget, Rationale: "Theme must not write raw SQL"},
			{Code: "sec.theme.noescape", Basket: domain.BasketBudget, Rationale: "Review |noescape allowlist"},
			{Code: "sec.theme.xss_reflected", Basket: domain.BasketFixTheme, Rationale: "Escape search/reflected output"},
			{Code: "sec.theme.xss_attr_breakout", Basket: domain.BasketFixTheme, Rationale: "Escape attributes in reflected output"},
			{Code: "sec.theme.path_disclosure", Basket: domain.BasketSiteDefaultOff, Rationale: "WP_DEBUG_DISPLAY off on prod"},
			{Code: "sec.theme.zip_open_failed", Basket: domain.BasketFixTheme, Rationale: "themeZip unreadable"},
			{Code: "sec.phpcs.ok", Basket: domain.BasketAccept, Rationale: "PHPCS Security clean"},
			{Code: "sec.phpcs.completed", Basket: domain.BasketAccept, Rationale: "PHPCS Security finished"},
			{Code: "sec.phpcs.security", Basket: domain.BasketFixTheme, Rationale: "Address PHPCS Security finding"},
			{Code: "sec.phpcs.zip_missing", Basket: domain.BasketBudget, Rationale: "themeZip required"},
			{Code: "sec.phpcs.exec_failed", Basket: domain.BasketFixTheme, Rationale: "PHPCS Security failed"},
			{Code: "sec.semgrep.ok", Basket: domain.BasketAccept, Rationale: "Semgrep clean"},
			{Code: "sec.semgrep.completed", Basket: domain.BasketAccept, Rationale: "Semgrep finished"},
			{Code: "sec.semgrep.finding", Basket: domain.BasketFixTheme, Rationale: "Address Semgrep theme rule"},
			{Code: "sec.semgrep.zip_missing", Basket: domain.BasketBudget, Rationale: "themeZip required"},
			{Code: "sec.semgrep.exec_failed", Basket: domain.BasketFixTheme, Rationale: "Semgrep failed"},
		}...)
	case "default":
		return defaultRules()
	default:
		return packRules("default")
	}
}

func defaultRules() []Rule {
	return []Rule{
		{Code: "demo.stub.ok", Basket: domain.BasketAccept, Rationale: "Demo stub informational finding"},
		{Code: "demo.stub.hint", Basket: domain.BasketBudget, Rationale: "Demo hint for future budgets"},
	}
}

// Map returns decisions for findings. Unmapped codes default to ACCEPT.
func (e *Engine) Map(findings []domain.Finding) []domain.Decision {
	out := make([]domain.Decision, 0, len(findings))
	seen := map[string]bool{}
	for _, f := range findings {
		if seen[f.Code] {
			continue
		}
		seen[f.Code] = true
		if r, ok := e.rules[f.Code]; ok {
			out = append(out, domain.Decision{
				FindingCode: f.Code,
				Basket:      r.Basket,
				Rationale:   r.Rationale,
			})
			continue
		}
		// Prefix heuristics for packs
		if d := heuristic(e.pack, f.Code); d != nil {
			out = append(out, *d)
			continue
		}
		out = append(out, domain.Decision{
			FindingCode: f.Code,
			Basket:      domain.BasketAccept,
			Rationale:   "unmapped finding; default ACCEPT",
		})
	}
	return out
}

func heuristic(pack, code string) *domain.Decision {
	switch {
	case pack == "wordpress-org" && (hasPrefix(code, "org.zip.missing_") ||
		hasPrefix(code, "org.zip.forbidden_") ||
		hasPrefix(code, "org.zip.screenshot_") ||
		hasPrefix(code, "org.zip.style_") ||
		hasPrefix(code, "org.zip.tag_") ||
		hasPrefix(code, "org.zip.resources_") ||
		hasPrefix(code, "org.zip.minified_") ||
		hasPrefix(code, "org.zip.nested_") ||
		hasPrefix(code, "org.zip.policy_")):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Theme packaging / Gate 1 issue"}
	case pack == "wordpress-org" && (code == "org.themecheck.required" || hasPrefix(code, "org.themecheck.") &&
		(hasSuffix(code, "_failed") || hasSuffix(code, "_missing") || code == "org.themecheck.wp_not_ready" || code == "org.themecheck.no_active_theme")):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Theme Check / Gate 2 issue"}
	case pack == "wordpress-org" && code == "org.themecheck.warning":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "Theme Check warning — review"}
	case pack == "wordpress-org" && (code == "org.matrix.status_5xx" || code == "org.matrix.status_unexpected" || code == "org.matrix.fetch_failed"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "HTTP smoke failure"}
	case pack == "wordpress-org" && code == "org.matrix.soft_404":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "Soft-404 — review template"}
	case pack == "wordpress-org" && code == "org.notice.found":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Theme PHP Notice/Warning/Deprecated under WP_DEBUG"}
	case pack == "wordpress-org" && hasPrefix(code, "org.keyboard.") &&
		code != "org.keyboard.ok" && code != "org.keyboard.summary" &&
		!hasSuffix(code, "_ok"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Keyboard / a11y chrome failure"}
	case pack == "lightspeed" && hasPrefix(code, "quality.axe.") && code != "quality.axe.ok":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Axe violation"}
	case pack == "lightspeed" && (code == "quality.vnu.error" || code == "quality.vnu.fetch_failed" || code == "quality.vnu.exec_failed" || code == "quality.vnu.parse_failed"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "HTML validation error"}
	case pack == "lightspeed" && code == "quality.vnu.soft_error":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "vnu softMode — content HTML"}
	case pack == "lightspeed" && (code == "quality.seo.og_title_missing" || code == "quality.seo.og_type_missing" ||
		code == "quality.seo.twitter_missing" || code == "quality.seo.jsonld_missing" || code == "quality.seo.jsonld_invalid"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "Social/JSON-LD profile gap"}
	case pack == "lightspeed" && hasPrefix(code, "quality.lighthouse.bytes_"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "Resource byte budget"}
	case pack == "lightspeed" && (code == "quality.css.parse_error" || code == "quality.css.forbidden" || code == "quality.css.exec_failed"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "CSS parse / forbidden rule"}
	case pack == "lightspeed" && code == "quality.css.no_files":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "No CSS scanned — review fixture/theme"}
	case pack == "lightspeed" && (code == "quality.seo.title_missing" || code == "quality.seo.viewport_missing" || code == "quality.seo.fetch_failed" || code == "quality.seo.parse_failed"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "SEO meta hard failure"}
	case pack == "lightspeed" && code == "quality.seo.h1":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "h1 count — review"}
	case pack == "lightspeed" && (hasPrefix(code, "quality.viewport.") && code != "quality.viewport.ok" ||
		hasPrefix(code, "quality.console.") && code != "quality.console.ok" ||
		code == "quality.extras.exec_failed" ||
		code == "quality.motion.unreduced" || code == "quality.motion.failed" ||
		code == "quality.links.broken" || code == "quality.links.fetch_failed" || code == "quality.links.failed"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Viewport / console / motion / links failure"}
	case pack == "lightspeed" && code == "quality.motion.emulation_failed":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "Reduced-motion emulation issue"}
	case pack == "lightspeed" && (code == "quality.lighthouse.lcp" || code == "quality.lighthouse.cls" ||
		code == "quality.lighthouse.tbt" || code == "quality.lighthouse.fcp" ||
		hasSuffix(code, ".missing") && hasPrefix(code, "quality.lighthouse.")):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketBudget, Rationale: "CWV budget — tune theme/site"}
	case pack == "lightspeed" && code == "quality.lighthouse.exec_failed":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Lighthouse execution failed"}
	case pack == "secure-baseline" && hasPrefix(code, "sec.recon."):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketSiteDefaultOff, Rationale: "Reduce attack surface"}
	}
	return nil
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

func hasSuffix(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}

// Pack returns the pack name.
func (e *Engine) Pack() string { return e.pack }
