package presets

import "github.com/fastygo/lab/packages/domain"

// Bindings override preset paths / adapter config for a SaaS job (F1.8).
// Relative paths stay as written; adapters resolve against repo root.
type Bindings struct {
	ThemeZip    string            // → adapter.config.themeZip (+ checks that already set themeZip)
	Root        string            // → adapter.config.root (static / static-web)
	BaseURL     string            // → adapter.config.baseUrl
	Config      map[string]string // merge into adapter.config
	CheckConfig map[string]string // merge into every check.config
}

// ApplyBindings mutates m in place.
func ApplyBindings(m *domain.Manifest, b Bindings) {
	if m == nil {
		return
	}
	if m.Spec.Adapter.Config == nil {
		m.Spec.Adapter.Config = map[string]string{}
	}
	for k, v := range b.Config {
		if v != "" {
			m.Spec.Adapter.Config[k] = v
		}
	}
	if b.ThemeZip != "" {
		m.Spec.Adapter.Config["themeZip"] = b.ThemeZip
	}
	if b.Root != "" {
		m.Spec.Adapter.Config["root"] = b.Root
	}
	if b.BaseURL != "" {
		m.Spec.Adapter.Config["baseUrl"] = b.BaseURL
	}
	for gi := range m.Spec.Gates {
		for ci := range m.Spec.Gates[gi].Checks {
			c := &m.Spec.Gates[gi].Checks[ci]
			if c.Config == nil {
				c.Config = map[string]string{}
			}
			if b.ThemeZip != "" {
				if _, ok := c.Config["themeZip"]; ok || needsThemeZip(c.Runner) {
					c.Config["themeZip"] = b.ThemeZip
				}
			}
			for k, v := range b.CheckConfig {
				if v != "" {
					c.Config[k] = v
				}
			}
		}
	}
}

func needsThemeZip(runner string) bool {
	switch runner {
	case "theme-check", "zip-lint", "composer-audit", "theme-sec", "phpcs-security", "semgrep":
		return true
	default:
		return false
	}
}
