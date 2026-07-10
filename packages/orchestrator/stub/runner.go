package stub

import (
	"context"
	"fmt"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner is an in-process demo runner (no Docker).
type Runner struct{}

func New() *Runner { return &Runner{} }

func (r *Runner) ID() string { return "stub" }

func (r *Runner) Run(_ context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	msg := req.Check.Config["message"]
	if msg == "" {
		msg = fmt.Sprintf("stub runner ok for check %s", req.Check.ID)
	}
	code := req.Check.Config["code"]
	if code == "" {
		code = "demo.stub.ok"
	}
	sev := domain.Severity(req.Check.Config["severity"])
	if sev == "" {
		sev = domain.SeverityInfo
	}
	return []domain.Finding{{
		Code:     code,
		Gate:     req.Gate,
		Check:    req.Check.ID,
		Severity: sev,
		Message:  msg,
		Target:   req.Target.BaseURL,
		Evidence: map[string]string{
			"runner": "stub",
			"urls":   fmt.Sprintf("%d", len(req.URLs)),
		},
	}}, nil
}
