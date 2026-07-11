package headers

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

var (
	generatorRe = regexp.MustCompile(`(?i)<meta\s+name=["']generator["']\s+content=["']([^"']+)["']`)
	indexOfRe   = regexp.MustCompile(`(?i)(<title>\s*Index of\s|/wp-content/[^<]*</a>)`)
)

// Runner checks security headers (S5) and recon probes (S1).
type Runner struct {
	client *http.Client
}

func New() *Runner {
	return &Runner{client: &http.Client{Timeout: 12 * time.Second, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= 5 {
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

	homeBody, homeHdr, homeErr := r.get(ctx, base+"/")
	if homeErr != nil {
		return []domain.Finding{{
			Code: "sec.headers.fetch_failed", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: homeErr.Error(), Target: base,
		}}, nil
	}

	findings = append(findings, r.checkHeaders(req, base, homeHdr)...)
	findings = append(findings, r.checkGenerator(req, base, homeBody)...)

	findings = append(findings, r.probeExposed(ctx, req, base+"/readme.html", "sec.recon.readme", domain.SeverityMedium, "WordPress readme.html exposed")...)
	findings = append(findings, r.probeExposed(ctx, req, base+"/license.txt", "sec.recon.license", domain.SeverityLow, "WordPress license.txt exposed")...)
	findings = append(findings, r.probeXMLRPC(ctx, req, base+"/xmlrpc.php")...)
	findings = append(findings, r.probeAuthorEnum(ctx, req, base)...)
	findings = append(findings, r.probeRESTUsers(ctx, req, base)...)
	findings = append(findings, r.probeRESTIndex(ctx, req, base)...)
	findings = append(findings, r.probeRegistration(ctx, req, base)...)
	findings = append(findings, r.probeDirListing(ctx, req, base+"/wp-content/uploads/", "sec.recon.dir_listing.uploads")...)
	findings = append(findings, r.probeDirListing(ctx, req, base+"/wp-content/themes/", "sec.recon.dir_listing.themes")...)
	findings = append(findings, r.probeSensitiveFiles(ctx, req, base)...)
	findings = append(findings, r.probeWPCron(ctx, req, base)...)
	findings = append(findings, r.probeFileEdit(ctx, req, base)...)

	if len(findings) == 0 {
		findings = append(findings, f(req, "sec.headers.ok", domain.SeverityInfo, "security headers/recon checks passed", base))
	}
	return findings, nil
}

func (r *Runner) checkHeaders(req ports.RunnerRequest, base string, hdr http.Header) []domain.Finding {
	var findings []domain.Finding
	if hdr.Get("X-Content-Type-Options") == "" {
		findings = append(findings, f(req, "sec.headers.nosniff", domain.SeverityMedium, "missing X-Content-Type-Options", base))
	}
	if hdr.Get("X-Frame-Options") == "" && !strings.Contains(strings.ToLower(hdr.Get("Content-Security-Policy")), "frame-ancestors") {
		findings = append(findings, f(req, "sec.headers.clickjacking", domain.SeverityMedium, "missing X-Frame-Options / CSP frame-ancestors", base))
	}
	if hdr.Get("Referrer-Policy") == "" {
		findings = append(findings, f(req, "sec.headers.referrer", domain.SeverityLow, "missing Referrer-Policy", base))
	}
	if hdr.Get("Content-Security-Policy") == "" && hdr.Get("Content-Security-Policy-Report-Only") == "" {
		findings = append(findings, f(req, "sec.headers.csp", domain.SeverityLow, "missing Content-Security-Policy (consider report-only first)", base))
	}
	if hdr.Get("Strict-Transport-Security") == "" {
		findings = append(findings, f(req, "sec.headers.hsts", domain.SeverityLow, "missing Strict-Transport-Security (prod HTTPS only)", base))
	}
	if hdr.Get("Permissions-Policy") == "" {
		findings = append(findings, f(req, "sec.headers.permissions", domain.SeverityLow, "missing Permissions-Policy", base))
	}
	return findings
}

func (r *Runner) checkGenerator(req ports.RunnerRequest, base, body string) []domain.Finding {
	m := generatorRe.FindStringSubmatch(body)
	if len(m) < 2 {
		return nil
	}
	return []domain.Finding{f(req, "sec.recon.generator", domain.SeverityLow,
		"WordPress generator meta exposes version: "+strings.TrimSpace(m[1]), base)}
}

func (r *Runner) get(ctx context.Context, rawURL string) (string, http.Header, error) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return "", nil, err
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	return string(b), resp.Header.Clone(), nil
}

func (r *Runner) getResp(ctx context.Context, rawURL string) (status int, body string, finalURL string, err error) {
	hreq, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return 0, "", "", err
	}
	resp, err := r.client.Do(hreq)
	if err != nil {
		return 0, "", "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))
	return resp.StatusCode, string(b), resp.Request.URL.String(), nil
}

func (r *Runner) probeExposed(ctx context.Context, req ports.RunnerRequest, rawURL, code string, sev domain.Severity, msg string) []domain.Finding {
	status, _, _, err := r.getResp(ctx, rawURL)
	if err != nil || status != 200 {
		return nil
	}
	return []domain.Finding{f(req, code, sev, msg, rawURL)}
}

func (r *Runner) probeXMLRPC(ctx context.Context, req ports.RunnerRequest, rawURL string) []domain.Finding {
	body := strings.NewReader(`<?xml version="1.0"?><methodCall><methodName>system.listMethods</methodName></methodCall>`)
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, body)
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
		return []domain.Finding{f(req, "sec.recon.xmlrpc", domain.SeverityHigh, "xmlrpc.php accepts system.listMethods", rawURL)}
	}
	return nil
}

func (r *Runner) probeAuthorEnum(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	status, _, finalURL, err := r.getResp(ctx, base+"/?author=1")
	if err != nil {
		return nil
	}
	u, _ := url.Parse(finalURL)
	path := ""
	if u != nil {
		path = u.Path
	}
	// Classic enum: redirect/rewrite to /author/<slug>/
	if strings.Contains(path, "/author/") && !strings.Contains(path, "author=1") {
		return []domain.Finding{f(req, "sec.recon.user_enum.author", domain.SeverityMedium,
			"author enumeration via ?author=1 → "+finalURL, base+"/?author=1")}
	}
	_ = status
	return nil
}

func (r *Runner) probeRESTUsers(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	status, body, _, err := r.getResp(ctx, base+"/wp-json/wp/v2/users")
	if err != nil || status != 200 {
		return nil
	}
	trim := strings.TrimSpace(body)
	if !strings.HasPrefix(trim, "[") && !strings.HasPrefix(trim, "{") {
		return nil
	}
	// Empty list is fine; non-empty user objects = enum
	var arr []map[string]any
	if err := json.Unmarshal([]byte(trim), &arr); err == nil && len(arr) > 0 {
		return []domain.Finding{f(req, "sec.recon.user_enum.rest", domain.SeverityMedium,
			"REST /wp/v2/users lists users without auth", base+"/wp-json/wp/v2/users")}
	}
	var one map[string]any
	if err := json.Unmarshal([]byte(trim), &one); err == nil {
		if _, ok := one["slug"]; ok {
			return []domain.Finding{f(req, "sec.recon.user_enum.rest", domain.SeverityMedium,
				"REST /wp/v2/users exposes user without auth", base+"/wp-json/wp/v2/users")}
		}
	}
	return nil
}

func (r *Runner) probeRESTIndex(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	status, body, _, err := r.getResp(ctx, base+"/wp-json/")
	if err != nil || status != 200 {
		return nil
	}
	if !strings.Contains(body, `"namespaces"`) && !strings.Contains(body, `"routes"`) {
		return nil
	}
	// Informational: Gutenberg needs REST; still worth noting open index.
	return []domain.Finding{f(req, "sec.recon.rest_index", domain.SeverityInfo,
		"REST API index is publicly readable", base+"/wp-json/")}
}

func (r *Runner) probeRegistration(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	status, body, _, err := r.getResp(ctx, base+"/wp-login.php?action=register")
	if err != nil || status != 200 {
		return nil
	}
	low := strings.ToLower(body)
	if strings.Contains(low, "registration is currently not allowed") ||
		strings.Contains(low, "registration disabled") {
		return nil
	}
	// Open registration form markers
	if strings.Contains(low, "user_login") && (strings.Contains(low, "user_email") || strings.Contains(low, "registerform")) {
		return []domain.Finding{f(req, "sec.recon.registration", domain.SeverityHigh,
			"user registration appears open", base+"/wp-login.php?action=register")}
	}
	return nil
}

func (r *Runner) probeDirListing(ctx context.Context, req ports.RunnerRequest, rawURL, code string) []domain.Finding {
	status, body, _, err := r.getResp(ctx, rawURL)
	if err != nil || status != 200 {
		return nil
	}
	if indexOfRe.MatchString(body) || strings.Contains(strings.ToLower(body), "index of /") {
		return []domain.Finding{f(req, code, domain.SeverityMedium, "directory listing enabled", rawURL)}
	}
	return nil
}

func (r *Runner) probeSensitiveFiles(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	paths := []struct {
		path string
		code string
		msg  string
	}{
		{"/.env", "sec.recon.sensitive.env", ".env exposed in webroot"},
		{"/wp-config.php.bak", "sec.recon.sensitive.wpconfig_bak", "wp-config.php.bak exposed"},
		{"/wp-config.bak", "sec.recon.sensitive.wpconfig_bak", "wp-config.bak exposed"},
		{"/wp-config.php~", "sec.recon.sensitive.wpconfig_bak", "wp-config.php~ exposed"},
		{"/wp-content/debug.log", "sec.recon.sensitive.debug_log", "debug.log exposed"},
		{"/composer.json", "sec.recon.sensitive.composer", "composer.json in webroot"},
		{"/wp-config.php.save", "sec.recon.sensitive.wpconfig_bak", "wp-config.php.save exposed"},
	}
	var findings []domain.Finding
	for _, p := range paths {
		status, body, _, err := r.getResp(ctx, base+p.path)
		if err != nil || status != 200 {
			continue
		}
		// Avoid false positives on HTML soft-404 themes
		if looksLikeHTMLSoft404(body) {
			continue
		}
		findings = append(findings, f(req, p.code, domain.SeverityHigh, p.msg, base+p.path))
	}
	return findings
}

func looksLikeHTMLSoft404(body string) bool {
	low := strings.ToLower(body)
	if strings.Contains(low, "<html") || strings.Contains(low, "<!doctype") {
		// Real .env / debug.log / bak are rarely full HTML pages
		if strings.Contains(low, "wordpress") || strings.Contains(low, "wp-content") {
			return true
		}
	}
	return false
}

func (r *Runner) probeWPCron(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	status, _, _, err := r.getResp(ctx, base+"/wp-cron.php")
	if err != nil {
		return nil
	}
	// wp-cron.php typically returns 200 with empty body when hit directly
	if status == 200 || status == 204 {
		return []domain.Finding{f(req, "sec.recon.wp_cron", domain.SeverityInfo,
			"wp-cron.php reachable unauthenticated (consider DISABLE_WP_CRON + system cron)", base+"/wp-cron.php")}
	}
	return nil
}

func (r *Runner) probeFileEdit(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	status, body, finalURL, err := r.getResp(ctx, base+"/wp-admin/theme-editor.php")
	if err != nil {
		return nil
	}
	low := strings.ToLower(body + finalURL)
	// Without auth we only see login; still recommend DISALLOW_FILE_EDIT on baseline.
	if status == 200 || status == 302 || strings.Contains(low, "wp-login") {
		return []domain.Finding{f(req, "sec.config.file_edit", domain.SeverityLow,
			"theme-editor.php endpoint present — ensure DISALLOW_FILE_EDIT on site baseline", base+"/wp-admin/theme-editor.php")}
	}
	return nil
}

func f(req ports.RunnerRequest, code string, sev domain.Severity, msg, target string) domain.Finding {
	return domain.Finding{Code: code, Gate: req.Gate, Check: req.Check.ID, Severity: sev, Message: msg, Target: target}
}
