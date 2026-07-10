package ports

import (
	"context"

	"github.com/fastygo/lab/packages/domain"
)

// TargetAdapter prepares and serves a target for runners.
type TargetAdapter interface {
	ID() string
	Capabilities() []string
	Prepare(ctx context.Context, config map[string]string) error
	Serve(ctx context.Context) (domain.Target, error)
	Matrix(ctx context.Context) ([]string, error)
	Teardown(ctx context.Context) error
}

// RunnerRequest is input to a check runner.
type RunnerRequest struct {
	Gate   string
	Check  domain.Check
	Target domain.Target
	URLs   []string
}

// Runner executes a check and returns findings.
type Runner interface {
	ID() string
	Run(ctx context.Context, req RunnerRequest) ([]domain.Finding, error)
}

// ArtifactStore persists reports (optional).
type ArtifactStore interface {
	SaveReport(ctx context.Context, report *domain.Report) error
}
