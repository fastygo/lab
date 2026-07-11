package themesec

import (
	"archive/zip"
	"context"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner scans theme zip for dangerous patterns (S4 static) and light dynamic XSS probes.
type Runner struct {
	client *http.Client
}

func New() *Runner {
	return &Runner{client: &http.Client{Timeout: 12 * time.Second}}
}

func (r *Runner) ID() string { return "theme-sec" }

var dangerPatterns = []struct {
	code string
	re   *regexp.Regexp
	sev  domain.Severity
	msg  string
}{
	{"sec.theme.eval", regexp.MustCompile(`(?i)\beval\s*\(`), domain.SeverityCritical, "eval() found"},
	{"sec.theme.create_function", regexp.MustCompile(`(?i)\bcreate_function\s*\(`), domain.SeverityCritical, "create_function() found"},
	{"sec.theme.assert", regexp.MustCompile(`(?i)\bassert\s*\(`), domain.SeverityHigh, "assert() found"},
	{"sec.theme.unserialize", regexp.MustCompile(`(?i)\bunserialize\s*\(`), domain.SeverityHigh, "unserialize() found"},
	{"sec.theme.shell", regexp.MustCompile(`(?i)\b(shell_exec|passthru|system|proc_open|popen)\s*\(`), domain.SeverityCritical, "shell execution function found"},
	// Shell backticks (Go RE2: no lookahead). PHPDoc `$var` ticks are skipped because $ is not allowed after `.
	{"sec.theme.backticks", regexp.MustCompile("\x60[a-zA-Z0-9_./-]+(?:\\s[^\x60]{0,80})?\x60"), domain.SeverityHigh, "backtick shell execution found"},
	{"sec.theme.file_get_remote", regexp.MustCompile(`(?i)file_get_contents\s*\(\s*['"]https?://`), domain.SeverityHigh, "file_get_contents remote URL"},
	{"sec.theme.curl_exec", regexp.MustCompile(`(?i)\bcurl_exec\s*\(`), domain.SeverityMedium, "curl_exec found — review SSRF"},
	{"sec.theme.open_redirect", regexp.MustCompile(`(?i)wp_redirect\s*\(\s*\$_(GET|POST|REQUEST)`), domain.SeverityHigh, "wp_redirect on user input"},
	{"sec.theme.echo_superglobal", regexp.MustCompile(`(?i)echo\s+\$_(GET|POST|REQUEST|COOKIE)`), domain.SeverityHigh, "echo of superglobal without escape"},
	{"sec.theme.sql_unprepared", regexp.MustCompile(`(?i)\$wpdb->(query|get_results|get_row|get_var)\s*\(\s*["'].*\$_(GET|POST|REQUEST)`), domain.SeverityCritical, "possible unprepared SQL with user input"},
	{"sec.theme.noescape", regexp.MustCompile(`\|noescape`), domain.SeverityMedium, "|noescape found — review allowlist (trusted WP HTML only)"},
}

func (r *Runner) Run(ctx context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	var findings []domain.Finding

	zipPath := firstNonEmpty(req.Check.Config["themeZip"], req.Target.Metadata["themeZip"])
	if zipPath != "" {
		findings = append(findings, r.scanZip(req, zipPath)...)
	} else {
		findings = append(findings, f(req, "sec.theme.zip_missing", domain.SeverityInfo,
			"no themeZip — static S4 scan skipped", req.Target.BaseURL))
	}

	base := strings.TrimRight(req.Target.BaseURL, "/")
	if base != "" {
		findings = append(findings, r.probeReflectedXSS(ctx, req, base)...)
		findings = append(findings, r.probeAttrBreakout(ctx, req, base)...)
		findings = append(findings, r.probeDebugLeak(ctx, req, base)...)
	}

	if len(findings) == 0 {
		findings = append(findings, f(req, "sec.theme.ok", domain.SeverityInfo, "theme security probes passed", base))
	}
	return findings, nil
}

func (r *Runner) scanZip(req ports.RunnerRequest, zipPath string) []domain.Finding {
	abs := zipPath
	if !filepath.IsAbs(abs) {
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
	}
	zr, err := zip.OpenReader(abs)
	if err != nil {
		return []domain.Finding{f(req, "sec.theme.zip_open_failed", domain.SeverityHigh, err.Error(), abs)}
	}
	defer zr.Close()

	hit := map[string]bool{}
	var findings []domain.Finding
	for _, zf := range zr.File {
		if zf.FileInfo().IsDir() {
			continue
		}
		name := strings.ReplaceAll(zf.Name, "\\", "/")
		lower := strings.ToLower(name)
		if strings.Contains(lower, "/vendor/") || strings.Contains(lower, "/node_modules/") {
			continue
		}
		if !strings.HasSuffix(lower, ".php") && !strings.HasSuffix(lower, ".inc") && !strings.HasSuffix(lower, ".latte") {
			continue
		}
		rc, err := zf.Open()
		if err != nil {
			continue
		}
		data, _ := io.ReadAll(io.LimitReader(rc, 512*1024))
		_ = rc.Close()
		body := string(data)
		for _, p := range dangerPatterns {
			if hit[p.code] {
				continue
			}
			isLatte := strings.HasSuffix(lower, ".latte")
			if isLatte && p.code != "sec.theme.noescape" {
				continue
			}
			if !isLatte && p.code == "sec.theme.noescape" {
				continue
			}
			if p.re.MatchString(body) {
				hit[p.code] = true
				findings = append(findings, f(req, p.code, p.sev, p.msg+" in "+name, abs))
			}
		}
	}
	if len(findings) == 0 {
		findings = append(findings, f(req, "sec.theme.static_ok", domain.SeverityInfo,
			"no banned dangerous PHP patterns in theme zip", abs))
	}
	return findings
}

func (r *Runner) probeReflectedXSS(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	payload := `<script>alert(1)</script>`
	u := base + "/?s=" + url.QueryEscape(payload)
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return []domain.Finding{f(req, "sec.theme.xss_fetch_failed", domain.SeverityMedium, err.Error(), u)}
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	s := string(body)
	// Unescaped script tag in HTML = fail
	if strings.Contains(s, "<script>alert(1)</script>") {
		return []domain.Finding{f(req, "sec.theme.xss_reflected", domain.SeverityCritical,
			"search query reflected unescaped as script tag", u)}
	}
	// Partially broken out
	if strings.Contains(s, "<script>") && strings.Contains(strings.ToLower(s), "alert(1)") &&
		!strings.Contains(s, "&lt;script&gt;") && !strings.Contains(s, "&#") {
		// Heuristic: script+alert without entity encoding nearby
		idx := strings.Index(s, "alert(1)")
		if idx >= 0 {
			window := s[max(0, idx-40):min(len(s), idx+40)]
			if strings.Contains(window, "<script") {
				return []domain.Finding{f(req, "sec.theme.xss_reflected", domain.SeverityCritical,
					"search query appears in script context unescaped", u)}
			}
		}
	}
	return []domain.Finding{f(req, "sec.theme.xss_search_ok", domain.SeverityInfo,
		"search reflected XSS probe did not find raw script payload", u)}
}

func (r *Runner) probeAttrBreakout(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	// Quote breakout in search reflection
	payload := `" onmouseover="alert(1)" x="`
	u := base + "/?s=" + url.QueryEscape(payload)
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	s := string(body)
	if strings.Contains(s, `onmouseover="alert(1)"`) || strings.Contains(s, "onmouseover='alert(1)'") {
		return []domain.Finding{f(req, "sec.theme.xss_attr_breakout", domain.SeverityCritical,
			"search query breaks out of attribute context", u)}
	}
	return []domain.Finding{f(req, "sec.theme.xss_attr_ok", domain.SeverityInfo,
		"attribute breakout probe did not find raw event handler", u)}
}

func (r *Runner) probeDebugLeak(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	u := base + "/?lab-debug-probe=1"
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	low := strings.ToLower(string(body))
	if strings.Contains(low, "wp_debug_display") || (strings.Contains(low, "/var/www/") && strings.Contains(low, ".php on line")) {
		return []domain.Finding{f(req, "sec.theme.path_disclosure", domain.SeverityMedium,
			"possible path/debug disclosure in HTML", u)}
	}
	return nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func f(req ports.RunnerRequest, code string, sev domain.Severity, msg, target string) domain.Finding {
	return domain.Finding{Code: code, Gate: req.Gate, Check: req.Check.ID, Severity: sev, Message: msg, Target: target}
}
