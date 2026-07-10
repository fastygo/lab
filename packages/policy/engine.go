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
	case "default":
		return []Rule{
			{
				Code:      "demo.stub.ok",
				Basket:    domain.BasketAccept,
				Rationale: "Demo stub informational finding",
			},
			{
				Code:      "demo.stub.hint",
				Basket:    domain.BasketBudget,
				Rationale: "Demo hint for future budgets",
			},
		}
	default:
		return packRules("default")
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
		out = append(out, domain.Decision{
			FindingCode: f.Code,
			Basket:      domain.BasketAccept,
			Rationale:   "unmapped finding; default ACCEPT",
		})
	}
	return out
}

// Pack returns the pack name.
func (e *Engine) Pack() string { return e.pack }
