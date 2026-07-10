package domain

import "testing"

func TestReportComputeStatus(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name      string
		findings  []Finding
		decisions []Decision
		want      ReportStatus
	}{
		{
			name:     "empty pass",
			findings: nil,
			want:     StatusPass,
		},
		{
			name: "info pass",
			findings: []Finding{
				{Code: "demo.ok", Severity: SeverityInfo},
			},
			want: StatusPass,
		},
		{
			name: "medium warn",
			findings: []Finding{
				{Code: "demo.med", Severity: SeverityMedium},
			},
			want: StatusWarn,
		},
		{
			name: "high fail",
			findings: []Finding{
				{Code: "demo.high", Severity: SeverityHigh},
			},
			want: StatusFail,
		},
		{
			name: "high accepted pass",
			findings: []Finding{
				{Code: "demo.high", Severity: SeverityHigh},
			},
			decisions: []Decision{
				{FindingCode: "demo.high", Basket: BasketAccept},
			},
			want: StatusPass,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			r := &Report{Findings: tc.findings, Decisions: tc.decisions}
			r.ComputeStatus()
			if r.Status != tc.want {
				t.Fatalf("status = %s, want %s", r.Status, tc.want)
			}
		})
	}
}

func TestReportSummarize(t *testing.T) {
	t.Parallel()
	r := &Report{Findings: []Finding{
		{Severity: SeverityCritical},
		{Severity: SeverityHigh},
		{Severity: SeverityInfo},
		{Severity: SeverityInfo},
	}}
	r.Summarize()
	if r.Summary.Total != 4 || r.Summary.Critical != 1 || r.Summary.High != 1 || r.Summary.Info != 2 {
		t.Fatalf("unexpected summary: %+v", r.Summary)
	}
}

func TestManifestValidate(t *testing.T) {
	t.Parallel()
	m := &Manifest{
		APIVersion: "lab.fastygo.dev/v1",
		Kind:       "LabManifest",
		Metadata:   ManifestMetadata{Name: "demo"},
		Spec: ManifestSpec{
			Lab:     "demo",
			Adapter: AdapterRef{ID: "noop"},
			Gates: []Gate{{
				ID: "G0",
				Checks: []Check{{
					ID:     "ping",
					Runner: "stub",
				}},
			}},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	m.Spec.Lab = ""
	if err := m.Validate(); err == nil {
		t.Fatal("expected error for empty lab")
	}
}
