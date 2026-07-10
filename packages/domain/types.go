package domain

import "time"

// Severity levels for findings.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// ReportStatus is the aggregate lab outcome.
type ReportStatus string

const (
	StatusPass ReportStatus = "pass"
	StatusWarn ReportStatus = "warn"
	StatusFail ReportStatus = "fail"
)

// Basket is a policy decision category.
type Basket string

const (
	BasketCutTarget      Basket = "CUT_TARGET"
	BasketFixTheme       Basket = "FIX_THEME"
	BasketFixSite        Basket = "FIX_SITE"
	BasketSiteDefaultOn  Basket = "SITE_DEFAULT_ON"
	BasketSiteDefaultOff Basket = "SITE_DEFAULT_OFF"
	BasketBudget         Basket = "BUDGET"
	BasketAccept         Basket = "ACCEPT"
	BasketBlockTag       Basket = "BLOCK_TAG"
)

// Target is the system under test.
type Target struct {
	BaseURL  string            `json:"baseUrl"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// Finding is one observation from a runner.
type Finding struct {
	Code     string            `json:"code"`
	Gate     string            `json:"gate"`
	Check    string            `json:"check"`
	Severity Severity          `json:"severity"`
	Message  string            `json:"message"`
	Evidence map[string]string `json:"evidence,omitempty"`
	Target   string            `json:"target,omitempty"`
}

// Decision is a policy mapping for a finding.
type Decision struct {
	FindingCode string `json:"findingCode"`
	Basket      Basket `json:"basket"`
	Rationale   string `json:"rationale"`
}

// Budget holds numeric thresholds for a gate/check.
type Budget struct {
	Name  string  `json:"name"`
	Min   float64 `json:"min,omitempty"`
	Max   float64 `json:"max,omitempty"`
	Value float64 `json:"value,omitempty"`
}

// Check is a single tool invocation inside a gate.
type Check struct {
	ID     string            `json:"id"`
	Runner string            `json:"runner"`
	Config map[string]string `json:"config,omitempty"`
}

// Gate groups checks.
type Gate struct {
	ID      string   `json:"id"`
	Checks  []Check  `json:"checks"`
	Budgets []Budget `json:"budgets,omitempty"`
}

// Summary counts findings by severity.
type Summary struct {
	Critical int `json:"critical"`
	High     int `json:"high"`
	Medium   int `json:"medium"`
	Low      int `json:"low"`
	Info     int `json:"info"`
	Total    int `json:"total"`
}

// Report is the lab run result.
type Report struct {
	Lab        string       `json:"lab"`
	Status     ReportStatus `json:"status"`
	StartedAt  time.Time    `json:"startedAt"`
	FinishedAt time.Time    `json:"finishedAt"`
	Findings   []Finding    `json:"findings"`
	Decisions  []Decision   `json:"decisions"`
	Summary    Summary      `json:"summary"`
}

// Summarize fills Summary from Findings.
func (r *Report) Summarize() {
	var s Summary
	for _, f := range r.Findings {
		s.Total++
		switch f.Severity {
		case SeverityCritical:
			s.Critical++
		case SeverityHigh:
			s.High++
		case SeverityMedium:
			s.Medium++
		case SeverityLow:
			s.Low++
		default:
			s.Info++
		}
	}
	r.Summary = s
}

// ComputeStatus sets Status from findings and decisions.
// critical/high findings whose decision basket is not ACCEPT cause fail.
// medium causes warn if status would otherwise be pass.
func (r *Report) ComputeStatus() {
	accepted := map[string]bool{}
	for _, d := range r.Decisions {
		if d.Basket == BasketAccept {
			accepted[d.FindingCode] = true
		}
	}
	status := StatusPass
	for _, f := range r.Findings {
		if accepted[f.Code] {
			continue
		}
		switch f.Severity {
		case SeverityCritical, SeverityHigh:
			r.Status = StatusFail
			return
		case SeverityMedium:
			status = StatusWarn
		}
	}
	r.Status = status
}
