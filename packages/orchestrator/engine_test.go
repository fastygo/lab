package orchestrator_test

import (
	"context"
	"testing"

	"github.com/fastygo/lab/packages/adapters/noop"
	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator"
	"github.com/fastygo/lab/packages/orchestrator/memory"
	"github.com/fastygo/lab/packages/orchestrator/stub"
)

func TestRunDemoManifest(t *testing.T) {
	t.Parallel()
	m := &domain.Manifest{
		APIVersion: "lab.fastygo.dev/v1",
		Kind:       "LabManifest",
		Metadata:   domain.ManifestMetadata{Name: "demo"},
		Spec: domain.ManifestSpec{
			Lab: "demo",
			Adapter: domain.AdapterRef{
				ID: "noop",
				Config: map[string]string{
					"baseUrl": "http://example.test",
				},
			},
			Policy: domain.PolicyRef{Pack: "default"},
			Gates: []domain.Gate{{
				ID: "G0-smoke",
				Checks: []domain.Check{
					{
						ID:     "ping",
						Runner: "stub",
						Config: map[string]string{
							"code":     "demo.stub.ok",
							"severity": "info",
							"message":  "demo ping",
						},
					},
					{
						ID:     "hint",
						Runner: "stub",
						Config: map[string]string{
							"code":     "demo.stub.hint",
							"severity": "info",
							"message":  "demo hint",
						},
					},
				},
			}},
		},
	}
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}

	store := memory.New()
	events := memory.NewEventSink()
	eng := orchestrator.New(
		[]orchestrator.TargetAdapter{noop.New()},
		[]orchestrator.Runner{stub.New()},
		store,
	).WithEvents(events).WithRunID("test-run-1")
	report, err := eng.Run(context.Background(), m)
	if err != nil {
		t.Fatal(err)
	}
	if report.Status != domain.StatusPass {
		t.Fatalf("status=%s", report.Status)
	}
	if report.Summary.Total != 2 {
		t.Fatalf("findings=%d", report.Summary.Total)
	}
	if len(store.Reports) != 1 {
		t.Fatalf("store reports=%d", len(store.Reports))
	}
	types := events.Types()
	wantOrder := []string{
		"run.started",
		"adapter.ready",
		"gate.started",
		"check.started",
		"check.finished",
		"check.started",
		"check.finished",
		"gate.finished",
		"run.finished",
	}
	if len(types) != len(wantOrder) {
		t.Fatalf("events=%v", types)
	}
	for i, w := range wantOrder {
		if types[i] != w {
			t.Fatalf("event[%d]=%s want %s full=%v", i, types[i], w, types)
		}
	}
	if events.Events[0].RunID != "test-run-1" {
		t.Fatalf("runId=%q", events.Events[0].RunID)
	}
}
