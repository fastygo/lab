package orchestrator

import (
	"context"
	"fmt"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
	"github.com/fastygo/lab/packages/policy"
)

// Runner is an alias for the port.
type Runner = ports.Runner

// TargetAdapter is an alias for the port.
type TargetAdapter = ports.TargetAdapter

// ArtifactStore is an alias for the port.
type ArtifactStore = ports.ArtifactStore

// Engine runs a lab manifest.
type Engine struct {
	adapters map[string]TargetAdapter
	runners  map[string]Runner
	store    ArtifactStore
}

// New creates an orchestrator engine.
func New(adapters []TargetAdapter, runners []Runner, store ArtifactStore) *Engine {
	e := &Engine{
		adapters: map[string]TargetAdapter{},
		runners:  map[string]Runner{},
		store:    store,
	}
	for _, a := range adapters {
		e.adapters[a.ID()] = a
	}
	for _, r := range runners {
		e.runners[r.ID()] = r
	}
	return e
}

// Run executes the manifest and returns a report.
func (e *Engine) Run(ctx context.Context, m *domain.Manifest) (*domain.Report, error) {
	started := time.Now().UTC()
	adapter, ok := e.adapters[m.Spec.Adapter.ID]
	if !ok {
		return nil, fmt.Errorf("unknown adapter %q", m.Spec.Adapter.ID)
	}
	if err := adapter.Prepare(ctx, m.Spec.Adapter.Config); err != nil {
		return nil, fmt.Errorf("adapter prepare: %w", err)
	}
	defer func() { _ = adapter.Teardown(ctx) }()

	target, err := adapter.Serve(ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter serve: %w", err)
	}
	urls, err := adapter.Matrix(ctx)
	if err != nil {
		return nil, fmt.Errorf("adapter matrix: %w", err)
	}

	var findings []domain.Finding
	for _, gate := range m.Spec.Gates {
		for _, check := range gate.Checks {
			runner, ok := e.runners[check.Runner]
			if !ok {
				return nil, fmt.Errorf("unknown runner %q (gate %s check %s)", check.Runner, gate.ID, check.ID)
			}
			got, err := runner.Run(ctx, ports.RunnerRequest{
				Gate:   gate.ID,
				Check:  check,
				Target: target,
				URLs:   urls,
			})
			if err != nil {
				return nil, fmt.Errorf("runner %s: %w", check.Runner, err)
			}
			findings = append(findings, got...)
		}
	}

	pack := m.Spec.Policy.Pack
	if pack == "" {
		pack = "default"
	}
	decisions := policy.NewEngine(pack).Map(findings)

	report := &domain.Report{
		Lab:        m.Spec.Lab,
		StartedAt:  started,
		FinishedAt: time.Now().UTC(),
		Findings:   findings,
		Decisions:  decisions,
	}
	report.Summarize()
	report.ComputeStatus()

	if e.store != nil {
		if err := e.store.SaveReport(ctx, report); err != nil {
			return nil, fmt.Errorf("save report: %w", err)
		}
	}
	return report, nil
}

// Labs returns registered lab ids discovered from... (CLI lists known labs statically).
func KnownLabs() []string {
	return []string{"demo"}
}
