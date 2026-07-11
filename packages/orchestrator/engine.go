package orchestrator

import (
	"context"
	"fmt"
	"strconv"
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

// EventSink is an alias for the port.
type EventSink = ports.EventSink

// Engine runs a lab manifest.
type Engine struct {
	adapters map[string]TargetAdapter
	runners  map[string]Runner
	store    ArtifactStore
	events   EventSink
	runID    string
}

// New creates an orchestrator engine.
func New(adapters []TargetAdapter, runners []Runner, store ArtifactStore) *Engine {
	e := &Engine{
		adapters: map[string]TargetAdapter{},
		runners:  map[string]Runner{},
		store:    store,
		events:   ports.NopSink{},
	}
	for _, a := range adapters {
		e.adapters[a.ID()] = a
	}
	for _, r := range runners {
		e.runners[r.ID()] = r
	}
	return e
}

// WithEvents sets the progress EventSink (Cycle F). Returns the same engine for chaining.
func (e *Engine) WithEvents(sink EventSink) *Engine {
	if sink == nil {
		e.events = ports.NopSink{}
		return e
	}
	e.events = sink
	return e
}

// WithRunID attaches an optional run id to emitted events (SaaS job id).
func (e *Engine) WithRunID(id string) *Engine {
	e.runID = id
	return e
}

// Run executes the manifest and returns a report.
func (e *Engine) Run(ctx context.Context, m *domain.Manifest) (*domain.Report, error) {
	started := time.Now().UTC()
	e.emit(ctx, domain.RunEvent{
		Type: domain.EventRunStarted,
		Lab:  m.Spec.Lab,
		Payload: map[string]string{
			"adapter": m.Spec.Adapter.ID,
		},
	})

	adapter, ok := e.adapters[m.Spec.Adapter.ID]
	if !ok {
		err := fmt.Errorf("unknown adapter %q", m.Spec.Adapter.ID)
		e.emitFailed(ctx, m.Spec.Lab, err)
		return nil, err
	}
	if err := adapter.Prepare(ctx, m.Spec.Adapter.Config); err != nil {
		err = fmt.Errorf("adapter prepare: %w", err)
		e.emitFailed(ctx, m.Spec.Lab, err)
		return nil, err
	}
	defer func() { _ = adapter.Teardown(ctx) }()

	target, err := adapter.Serve(ctx)
	if err != nil {
		err = fmt.Errorf("adapter serve: %w", err)
		e.emitFailed(ctx, m.Spec.Lab, err)
		return nil, err
	}
	e.emit(ctx, domain.RunEvent{
		Type:    domain.EventAdapterReady,
		Lab:     m.Spec.Lab,
		Adapter: m.Spec.Adapter.ID,
		BaseURL: target.BaseURL,
	})

	urls, err := adapter.Matrix(ctx)
	if err != nil {
		err = fmt.Errorf("adapter matrix: %w", err)
		e.emitFailed(ctx, m.Spec.Lab, err)
		return nil, err
	}

	var findings []domain.Finding
	for _, gate := range m.Spec.Gates {
		e.emit(ctx, domain.RunEvent{Type: domain.EventGateStarted, Lab: m.Spec.Lab, Gate: gate.ID})
		gateStart := time.Now()
		gateFindings := 0

		// Refresh matrix before each gate so seed from earlier gates (e.g. theme-check) expands URLs.
		if refreshed, err := adapter.Matrix(ctx); err == nil {
			urls = refreshed
		}
		for _, check := range gate.Checks {
			runner, ok := e.runners[check.Runner]
			if !ok {
				err := fmt.Errorf("unknown runner %q (gate %s check %s)", check.Runner, gate.ID, check.ID)
				e.emitFailed(ctx, m.Spec.Lab, err)
				return nil, err
			}
			e.emit(ctx, domain.RunEvent{
				Type:   domain.EventCheckStarted,
				Lab:    m.Spec.Lab,
				Gate:   gate.ID,
				Check:  check.ID,
				Runner: check.Runner,
			})
			checkStart := time.Now()
			got, err := runner.Run(ctx, ports.RunnerRequest{
				Gate:   gate.ID,
				Check:  check,
				Target: target,
				URLs:   urls,
			})
			if err != nil {
				err = fmt.Errorf("runner %s: %w", check.Runner, err)
				e.emitFailed(ctx, m.Spec.Lab, err)
				return nil, err
			}
			findings = append(findings, got...)
			gateFindings += len(got)
			e.emit(ctx, domain.RunEvent{
				Type:   domain.EventCheckFinished,
				Lab:    m.Spec.Lab,
				Gate:   gate.ID,
				Check:  check.ID,
				Runner: check.Runner,
				Payload: map[string]string{
					"findingCount": strconv.Itoa(len(got)),
					"durationMs":   strconv.FormatInt(time.Since(checkStart).Milliseconds(), 10),
				},
			})
		}
		e.emit(ctx, domain.RunEvent{
			Type: domain.EventGateFinished,
			Lab:  m.Spec.Lab,
			Gate: gate.ID,
			Payload: map[string]string{
				"findingCount": strconv.Itoa(gateFindings),
				"durationMs":   strconv.FormatInt(time.Since(gateStart).Milliseconds(), 10),
			},
		})
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

	e.emit(ctx, domain.RunEvent{
		Type:   domain.EventRunFinished,
		Lab:    m.Spec.Lab,
		Status: report.Status,
		Payload: map[string]string{
			"total":      strconv.Itoa(report.Summary.Total),
			"critical":   strconv.Itoa(report.Summary.Critical),
			"high":       strconv.Itoa(report.Summary.High),
			"medium":     strconv.Itoa(report.Summary.Medium),
			"durationMs": strconv.FormatInt(report.FinishedAt.Sub(report.StartedAt).Milliseconds(), 10),
		},
	})

	if e.store != nil {
		if err := e.store.SaveReport(ctx, report); err != nil {
			return nil, fmt.Errorf("save report: %w", err)
		}
	}
	return report, nil
}

func (e *Engine) emit(ctx context.Context, ev domain.RunEvent) {
	if e.events == nil {
		return
	}
	if ev.TS.IsZero() {
		ev.TS = time.Now().UTC()
	}
	if ev.RunID == "" {
		ev.RunID = e.runID
	}
	_ = e.events.Emit(ctx, ev)
}

func (e *Engine) emitFailed(ctx context.Context, lab string, err error) {
	e.emit(ctx, domain.RunEvent{
		Type:  domain.EventRunFailed,
		Lab:   lab,
		Error: err.Error(),
	})
}

// KnownLabs returns product lab ids (also exposed via packages/registry).
func KnownLabs() []string {
	return []string{"demo", "quality", "org", "sec", "static-web"}
}
