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
			{Code: "sec.recon.xmlrpc", Basket: domain.BasketSiteDefaultOff, Rationale: "Disable xmlrpc on site baseline"},
			{Code: "sec.recon.readme", Basket: domain.BasketSiteDefaultOff, Rationale: "Remove readme.html from deploy"},
			{Code: "sec.wpscan.completed", Basket: domain.BasketAccept, Rationale: "Enumeration completed"},
			{Code: "runner.docker.unavailable", Basket: domain.BasketAccept, Rationale: "WPScan needs Docker"},
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
		code == "quality.extras.exec_failed"):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Viewport / console failure"}
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
