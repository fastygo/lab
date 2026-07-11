// Package worker executes queued lab runs with the shared orchestrator (Cycle F).
package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator"
	"github.com/fastygo/lab/packages/registry"
	"github.com/fastygo/lab/packages/runstore"
)

// StoreSink writes orchestrator events into a runstore.
type StoreSink struct {
	Store runstore.Store
	RunID string
}

func (s *StoreSink) Emit(ctx context.Context, ev domain.RunEvent) error {
	if s == nil || s.Store == nil || s.RunID == "" {
		return nil
	}
	ev.RunID = s.RunID
	return s.Store.AppendEvent(ctx, s.RunID, ev)
}

// Options configure a worker execution.
type Options struct {
	Store    runstore.Store
	RepoRoot string
}

// Execute loads the run's manifest snapshot and runs the orchestrator.
func Execute(ctx context.Context, opts Options, runID string) error {
	if opts.Store == nil {
		return fmt.Errorf("worker: store required")
	}
	run, err := opts.Store.GetRun(ctx, runID)
	if err != nil {
		return err
	}
	if len(run.ManifestJSON) == 0 {
		return failRun(ctx, opts.Store, run, fmt.Errorf("run has empty manifest"))
	}

	m, err := domain.ParseManifest(run.ManifestJSON)
	if err != nil {
		return failRun(ctx, opts.Store, run, fmt.Errorf("parse manifest: %w", err))
	}

	now := time.Now().UTC()
	run.Status = runstore.StatusRunning
	run.StartedAt = &now
	run.Error = ""
	if err := opts.Store.UpdateRun(ctx, run); err != nil {
		return err
	}

	root := opts.RepoRoot
	if root == "" {
		root = registry.FindRepoRoot()
	}
	eng := orchestrator.New(
		registry.DefaultAdapters(root),
		registry.DefaultRunners(),
		nil,
	).WithEvents(&StoreSink{Store: opts.Store, RunID: runID}).WithRunID(runID)

	report, err := eng.Run(ctx, m)
	finished := time.Now().UTC()
	run.FinishedAt = &finished
	if err != nil {
		run.Status = runstore.StatusError
		run.Error = err.Error()
		_ = opts.Store.UpdateRun(ctx, run)
		return err
	}
	run.Report = report
	switch report.Status {
	case domain.StatusPass:
		run.Status = runstore.StatusPass
	case domain.StatusWarn:
		run.Status = runstore.StatusWarn
	default:
		run.Status = runstore.StatusFail
	}
	return opts.Store.UpdateRun(ctx, run)
}

func failRun(ctx context.Context, store runstore.Store, run *runstore.Run, err error) error {
	finished := time.Now().UTC()
	run.Status = runstore.StatusError
	run.Error = err.Error()
	run.FinishedAt = &finished
	_ = store.UpdateRun(ctx, run)
	return err
}
