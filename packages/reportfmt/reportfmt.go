// Package reportfmt formats and compares lab Reports (Cycle F3.5 / F3.6).
package reportfmt

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/runstore"
)

// Diff is a regression-oriented comparison of two reports.
type Diff struct {
	BaseID   string          `json:"baseId"`
	HeadID   string          `json:"headId"`
	BaseLab  string          `json:"baseLab"`
	HeadLab  string          `json:"headLab"`
	BaseStatus string        `json:"baseStatus"`
	HeadStatus string        `json:"headStatus"`
	BaseSummary domain.Summary `json:"baseSummary"`
	HeadSummary domain.Summary `json:"headSummary"`
	Added    []domain.Finding `json:"added"`    // in head, not base (by code+target)
	Removed  []domain.Finding `json:"removed"`  // in base, not head
	Changed  []FindingChange  `json:"changed"`  // same key, different severity/message
	SameCount int             `json:"sameCount"`
}

// FindingChange is a finding present in both runs with different severity or message.
type FindingChange struct {
	Key      string         `json:"key"`
	Base     domain.Finding `json:"base"`
	Head     domain.Finding `json:"head"`
}

func findingKey(f domain.Finding) string {
	return f.Code + "|" + f.Gate + "|" + f.Check + "|" + f.Target
}

// CompareReports diffs two reports (base → head).
func CompareReports(baseID, headID string, base, head *domain.Report) Diff {
	d := Diff{BaseID: baseID, HeadID: headID}
	if base != nil {
		d.BaseLab = base.Lab
		d.BaseStatus = string(base.Status)
		d.BaseSummary = base.Summary
	}
	if head != nil {
		d.HeadLab = head.Lab
		d.HeadStatus = string(head.Status)
		d.HeadSummary = head.Summary
	}
	baseMap := map[string]domain.Finding{}
	headMap := map[string]domain.Finding{}
	if base != nil {
		for _, f := range base.Findings {
			if f.Severity == domain.SeverityInfo {
				continue
			}
			baseMap[findingKey(f)] = f
		}
	}
	if head != nil {
		for _, f := range head.Findings {
			if f.Severity == domain.SeverityInfo {
				continue
			}
			headMap[findingKey(f)] = f
		}
	}
	for k, hf := range headMap {
		bf, ok := baseMap[k]
		if !ok {
			d.Added = append(d.Added, hf)
			continue
		}
		if bf.Severity != hf.Severity || bf.Message != hf.Message {
			d.Changed = append(d.Changed, FindingChange{Key: k, Base: bf, Head: hf})
		} else {
			d.SameCount++
		}
	}
	for k, bf := range baseMap {
		if _, ok := headMap[k]; !ok {
			d.Removed = append(d.Removed, bf)
		}
	}
	sort.Slice(d.Added, func(i, j int) bool { return d.Added[i].Code < d.Added[j].Code })
	sort.Slice(d.Removed, func(i, j int) bool { return d.Removed[i].Code < d.Removed[j].Code })
	sort.Slice(d.Changed, func(i, j int) bool { return d.Changed[i].Key < d.Changed[j].Key })
	return d
}

// Markdown renders a run report as markdown.
func Markdown(run *runstore.Run, report *domain.Report) string {
	var b strings.Builder
	lab := ""
	status := string(run.Status)
	if report != nil {
		lab = report.Lab
		if report.Status != "" {
			status = string(report.Status)
		}
	} else {
		lab = run.Lab
	}
	fmt.Fprintf(&b, "# Lab report — %s\n\n", lab)
	fmt.Fprintf(&b, "- **Run ID:** `%s`\n", run.ID)
	fmt.Fprintf(&b, "- **Status:** %s\n", status)
	if run.StartedAt != nil {
		fmt.Fprintf(&b, "- **Started:** %s\n", run.StartedAt.UTC().Format(time.RFC3339))
	}
	if run.FinishedAt != nil {
		fmt.Fprintf(&b, "- **Finished:** %s\n", run.FinishedAt.UTC().Format(time.RFC3339))
	}
	if run.Error != "" {
		fmt.Fprintf(&b, "- **Error:** %s\n", run.Error)
	}
	b.WriteString("\n")
	if report == nil {
		b.WriteString("_Report not ready._\n")
		return b.String()
	}
	s := report.Summary
	fmt.Fprintf(&b, "## Summary\n\n| Severity | Count |\n|----------|------:|\n")
	fmt.Fprintf(&b, "| critical | %d |\n| high | %d |\n| medium | %d |\n| low | %d |\n| info | %d |\n| **total** | %d |\n\n",
		s.Critical, s.High, s.Medium, s.Low, s.Info, s.Total)

	if len(report.Decisions) > 0 {
		baskets := map[string]int{}
		for _, d := range report.Decisions {
			baskets[string(d.Basket)]++
		}
		keys := make([]string, 0, len(baskets))
		for k := range baskets {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		b.WriteString("## Decision baskets\n\n")
		for _, k := range keys {
			fmt.Fprintf(&b, "- `%s`: %d\n", k, baskets[k])
		}
		b.WriteString("\n")
	}

	b.WriteString("## Findings\n\n")
	b.WriteString("| Sev | Code | Gate | Message |\n|-----|------|------|----------|\n")
	for _, f := range report.Findings {
		if f.Severity == domain.SeverityInfo {
			continue
		}
		msg := strings.ReplaceAll(f.Message, "|", "\\|")
		msg = strings.ReplaceAll(msg, "\n", " ")
		fmt.Fprintf(&b, "| %s | `%s` | %s | %s |\n", f.Severity, f.Code, f.Gate, msg)
	}
	return b.String()
}

// DiffMarkdown renders a compare result as markdown.
func DiffMarkdown(d Diff) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Compare runs\n\n")
	fmt.Fprintf(&b, "- **Base:** `%s` (%s) — %s\n", d.BaseID, d.BaseLab, d.BaseStatus)
	fmt.Fprintf(&b, "- **Head:** `%s` (%s) — %s\n\n", d.HeadID, d.HeadLab, d.HeadStatus)
	fmt.Fprintf(&b, "## Summary delta\n\n| | Base | Head |\n|--|-----:|-----:|\n")
	fmt.Fprintf(&b, "| high | %d | %d |\n| medium | %d | %d |\n| total | %d | %d |\n\n",
		d.BaseSummary.High, d.HeadSummary.High,
		d.BaseSummary.Medium, d.HeadSummary.Medium,
		d.BaseSummary.Total, d.HeadSummary.Total)
	fmt.Fprintf(&b, "- Unchanged (non-info): %d\n", d.SameCount)
	fmt.Fprintf(&b, "- Added: %d · Removed: %d · Changed: %d\n\n", len(d.Added), len(d.Removed), len(d.Changed))

	writeFindings := func(title string, list []domain.Finding) {
		if len(list) == 0 {
			return
		}
		fmt.Fprintf(&b, "## %s\n\n", title)
		for _, f := range list {
			fmt.Fprintf(&b, "- **%s** `%s` (%s) — %s\n", f.Severity, f.Code, f.Gate, f.Message)
		}
		b.WriteString("\n")
	}
	writeFindings("Added in head", d.Added)
	writeFindings("Removed since base", d.Removed)
	if len(d.Changed) > 0 {
		b.WriteString("## Changed\n\n")
		for _, c := range d.Changed {
			fmt.Fprintf(&b, "- `%s`: %s → %s — %s\n", c.Base.Code, c.Base.Severity, c.Head.Severity, c.Head.Message)
		}
	}
	return b.String()
}
