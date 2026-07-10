package httpmatrix

import (
	"context"
	"fmt"
	"strings"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner records the adapter URL matrix as informational findings (C3 minimal).
type Runner struct{}

func New() *Runner { return &Runner{} }

func (r *Runner) ID() string { return "http-matrix" }

func (r *Runner) Run(_ context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	if len(req.URLs) == 0 {
		return []domain.Finding{{
			Code:     "org.matrix.empty",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityMedium,
			Message:  "URL matrix is empty",
			Target:   req.Target.BaseURL,
		}}, nil
	}
	return []domain.Finding{{
		Code:     "org.matrix.listed",
		Gate:     req.Gate,
		Check:    req.Check.ID,
		Severity: domain.SeverityInfo,
		Message:  fmt.Sprintf("matrix has %d URLs", len(req.URLs)),
		Target:   req.Target.BaseURL,
		Evidence: map[string]string{"urls": strings.Join(req.URLs, " ")},
	}}, nil
}
