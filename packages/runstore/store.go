// Package runstore persists lab jobs, reports, and progress events (Cycle F).
package runstore

import (
	"context"
	"time"

	"github.com/fastygo/lab/packages/domain"
)

// RunStatus is the SaaS job lifecycle (broader than ReportStatus).
type RunStatus string

const (
	StatusQueued  RunStatus = "queued"
	StatusRunning RunStatus = "running"
	StatusPass    RunStatus = "pass"
	StatusWarn    RunStatus = "warn"
	StatusFail    RunStatus = "fail"
	StatusError   RunStatus = "error"
)

// Run is one lab job.
type Run struct {
	ID           string          `json:"id"`
	Lab          string          `json:"lab"`
	Status       RunStatus       `json:"status"`
	ManifestJSON []byte          `json:"-"`
	Report       *domain.Report  `json:"report,omitempty"`
	Error        string          `json:"error,omitempty"`
	CreatedAt    time.Time       `json:"createdAt"`
	StartedAt    *time.Time      `json:"startedAt,omitempty"`
	FinishedAt   *time.Time      `json:"finishedAt,omitempty"`
}

// Store persists runs and events.
type Store interface {
	CreateRun(ctx context.Context, run *Run) error
	GetRun(ctx context.Context, id string) (*Run, error)
	ListRuns(ctx context.Context, lab string, limit int) ([]*Run, error)
	UpdateRun(ctx context.Context, run *Run) error
	AppendEvent(ctx context.Context, runID string, ev domain.RunEvent) error
	ListEvents(ctx context.Context, runID string) ([]domain.RunEvent, error)
}
