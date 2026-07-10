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
	case pack == "lightspeed" && hasPrefix(code, "quality.axe.") && code != "quality.axe.ok":
		return &domain.Decision{FindingCode: code, Basket: domain.BasketFixTheme, Rationale: "Axe violation"}
	case pack == "secure-baseline" && hasPrefix(code, "sec.recon."):
		return &domain.Decision{FindingCode: code, Basket: domain.BasketSiteDefaultOff, Rationale: "Reduce attack surface"}
	}
	return nil
}

func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

// Pack returns the pack name.
func (e *Engine) Pack() string { return e.pack }
