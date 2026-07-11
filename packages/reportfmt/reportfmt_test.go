package reportfmt_test

import (
	"strings"
	"testing"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/reportfmt"
	"github.com/fastygo/lab/packages/runstore"
)

func TestCompareAndMarkdown(t *testing.T) {
	t.Parallel()
	base := &domain.Report{
		Lab: "quality", Status: domain.StatusFail,
		Summary: domain.Summary{High: 1, Total: 2},
		Findings: []domain.Finding{
			{Code: "a.x", Gate: "Q1", Severity: domain.SeverityHigh, Message: "old"},
			{Code: "b.y", Gate: "Q2", Severity: domain.SeverityMedium, Message: "keep"},
		},
	}
	head := &domain.Report{
		Lab: "quality", Status: domain.StatusWarn,
		Summary: domain.Summary{High: 0, Medium: 2, Total: 2},
		Findings: []domain.Finding{
			{Code: "a.x", Gate: "Q1", Severity: domain.SeverityMedium, Message: "new"},
			{Code: "c.z", Gate: "Q3", Severity: domain.SeverityMedium, Message: "added"},
		},
	}
	d := reportfmt.CompareReports("base", "head", base, head)
	if len(d.Added) != 1 || d.Added[0].Code != "c.z" {
		t.Fatalf("added=%v", d.Added)
	}
	if len(d.Removed) != 1 || d.Removed[0].Code != "b.y" {
		t.Fatalf("removed=%v", d.Removed)
	}
	if len(d.Changed) != 1 {
		t.Fatalf("changed=%v", d.Changed)
	}
	md := reportfmt.DiffMarkdown(d)
	if !strings.Contains(md, "Compare runs") || !strings.Contains(md, "c.z") {
		t.Fatalf("md=%s", md)
	}
	now := time.Now().UTC()
	run := &runstore.Run{ID: "r1", Lab: "quality", Status: runstore.StatusFail, StartedAt: &now, FinishedAt: &now}
	out := reportfmt.Markdown(run, head)
	if !strings.Contains(out, "# Lab report") || !strings.Contains(out, "a.x") {
		t.Fatalf("report md=%s", out)
	}
}
