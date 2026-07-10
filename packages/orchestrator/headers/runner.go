package headers

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner checks security headers and light recon probes.
type Runner struct {
	client *http.Client
}

func New() *Runner {
	return &Runner{client: &http.Client{Timeout: 8 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 3 {
			return http.ErrUseLastResponse
		}
		return nil
	}}}
}

func (r *Runner) ID() string { return "headers" }

func (r *Runner) Run(ctx context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	base := strings.TrimRight(req.Target.BaseURL, "/")
	if base == "" {
		return []domain.Finding{{
			Code: "sec.headers.no_target", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: "empty target URL",
		}}, nil
	}

	var findings []domain.Finding
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, base+"/", nil)
	if err != nil {
		return nil, err
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return []domain.Finding{{
			Code: "sec.headers.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: err.Error(), Target: base,
		}}, nil
	}
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	hdr := resp.Header
	if hdr.Get("X-Content-Type-Options") == "" {
		findings = append(findings, f(req, "sec.headers.nosniff", domain.SeverityMedium, "missing X-Content-Type-Options", base))
	}
	if hdr.Get("X-Frame-Options") == "" && !strings.Contains(strings.ToLower(hdr.Get("Content-Security-Policy")), "frame-ancestors") {
		findings = append(findings, f(req, "sec.headers.clickjacking", domain.SeverityMedium, "missing X-Frame-Options / CSP frame-ancestors", base))
	}
	if hdr.Get("Referrer-Policy") == "" {
		findings = append(findings, f(req, "sec.headers.referrer", domain.SeverityLow, "missing Referrer-Policy", base))
	}

	findings = append(findings, r.probe(ctx, req, base+"/readme.html", "sec.recon.readme", "WordPress readme.html exposed")...)
	findings = append(findings, r.probeXMLRPC(ctx, req, base+"/xmlrpc.php")...)

	if len(findings) == 0 {
		findings = append(findings, f(req, "sec.headers.ok", domain.SeverityInfo, "security headers/recon checks passed", base))
	}
	return findings, nil
}

func (r *Runner) probe(ctx context.Context, req ports.RunnerRequest, url, code, msg string) []domain.Finding {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode == 200 {
		return []domain.Finding{f(req, code, domain.SeverityMedium, msg, url)}
	}
	return nil
}

func (r *Runner) probeXMLRPC(ctx context.Context, req ports.RunnerRequest, url string) []domain.Finding {
	body := strings.NewReader(`<?xml version="1.0"?><methodCall><methodName>system.listMethods</methodName></methodCall>`)
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	if err != nil {
		return nil
	}
	hreq.Header.Set("Content-Type", "text/xml")
	resp, err := r.client.Do(hreq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode == 200 && strings.Contains(string(b), "methodResponse") {
		return []domain.Finding{f(req, "sec.recon.xmlrpc", domain.SeverityHigh, "xmlrpc.php accepts system.listMethods", url)}
	}
	return nil
}

func f(req ports.RunnerRequest, code string, sev domain.Severity, msg, target string) domain.Finding {
	return domain.Finding{Code: code, Gate: req.Gate, Check: req.Check.ID, Severity: sev, Message: msg, Target: target}
}
