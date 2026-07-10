package httpmatrix

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner GETs each matrix URL and asserts HTTP status (Gate 3 smoke).
type Runner struct {
	client *http.Client
}

func New() *Runner {
	return &Runner{client: &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}}
}

func (r *Runner) ID() string { return "http-matrix" }

func (r *Runner) Run(ctx context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	urls := req.URLs
	if len(urls) == 0 {
		return []domain.Finding{{
			Code:     "org.matrix.empty",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityMedium,
			Message:  "URL matrix is empty",
			Target:   req.Target.BaseURL,
		}}, nil
	}

	listOnly := req.Check.Config["listOnly"] == "true"
	if listOnly {
		return []domain.Finding{{
			Code:     "org.matrix.listed",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityInfo,
			Message:  fmt.Sprintf("matrix has %d URLs", len(urls)),
			Target:   req.Target.BaseURL,
			Evidence: map[string]string{"urls": strings.Join(urls, " ")},
		}}, nil
	}

	okExpect := 200
	if v := req.Check.Config["okStatus"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			okExpect = n
		}
	}

	var findings []domain.Finding
	okCount := 0
	for _, u := range urls {
		f, ok := r.probe(ctx, req, u, okExpect)
		findings = append(findings, f...)
		if ok {
			okCount++
		}
	}
	findings = append(findings, domain.Finding{
		Code:     "org.matrix.smoke_summary",
		Gate:     req.Gate,
		Check:    req.Check.ID,
		Severity: domain.SeverityInfo,
		Message:  fmt.Sprintf("HTTP smoke: %d/%d URLs ok", okCount, len(urls)),
		Target:   req.Target.BaseURL,
		Evidence: map[string]string{
			"ok":    strconv.Itoa(okCount),
			"total": strconv.Itoa(len(urls)),
		},
	})
	return findings, nil
}

func (r *Runner) probe(ctx context.Context, req ports.RunnerRequest, url string, okExpect int) ([]domain.Finding, bool) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return []domain.Finding{{
			Code: "org.matrix.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: url,
		}}, false
	}
	hreq.Header.Set("User-Agent", "FastyGo-Lab-HTTP-Smoke/0.2")
	resp, err := r.client.Do(hreq)
	if err != nil {
		return []domain.Finding{{
			Code: "org.matrix.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: url,
		}}, false
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	status := resp.StatusCode

	// 404 URLs in matrix may intentionally expect 404
	expect404 := strings.Contains(url, "does-not-exist") || strings.Contains(url, "lab-404")
	if expect404 {
		if status == 404 || status == 200 {
			// WP often soft-404s with 200 — warn
			if status == 200 {
				return []domain.Finding{{
					Code: "org.matrix.soft_404", Gate: req.Gate, Check: req.Check.ID,
					Severity: domain.SeverityMedium,
					Message:  "expected 404 URL returned 200 (possible soft-404)",
					Target:   url,
					Evidence: map[string]string{"status": strconv.Itoa(status)},
				}}, true
			}
			return []domain.Finding{{
				Code: "org.matrix.ok", Gate: req.Gate, Check: req.Check.ID,
				Severity: domain.SeverityInfo, Message: "404 URL returned 404", Target: url,
				Evidence: map[string]string{"status": "404"},
			}}, true
		}
	}

	if status >= 500 {
		return []domain.Finding{{
			Code: "org.matrix.status_5xx", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh,
			Message:  fmt.Sprintf("HTTP %d for %s", status, url),
			Target:   url,
			Evidence: map[string]string{"status": strconv.Itoa(status), "snippet": clip(string(body), 200)},
		}}, false
	}
	if status != okExpect && status != 301 && status != 302 && status != 303 && status != 307 && status != 308 {
		sev := domain.SeverityMedium
		if status == 404 {
			sev = domain.SeverityHigh
		}
		return []domain.Finding{{
			Code: "org.matrix.status_unexpected", Gate: req.Gate, Check: req.Check.ID,
			Severity: sev,
			Message:  fmt.Sprintf("HTTP %d (want %d) for %s", status, okExpect, url),
			Target:   url,
			Evidence: map[string]string{"status": strconv.Itoa(status)},
		}}, false
	}
	return []domain.Finding{{
		Code: "org.matrix.ok", Gate: req.Gate, Check: req.Check.ID,
		Severity: domain.SeverityInfo,
		Message:  fmt.Sprintf("HTTP %d OK", status),
		Target:   url,
		Evidence: map[string]string{"status": strconv.Itoa(status)},
	}}, true
}

func clip(s string, n int) string {
	s = strings.TrimSpace(s)
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
