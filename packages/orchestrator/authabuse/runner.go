package authabuse

import (
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// Runner runs lab-only auth abuse probes (S3).
type Runner struct{}

func New() *Runner { return &Runner{} }

func (r *Runner) ID() string { return "auth-abuse" }

func (r *Runner) Run(ctx context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	base := strings.TrimRight(req.Target.BaseURL, "/")
	if base == "" {
		return []domain.Finding{{
			Code: "sec.auth.no_target", Gate: req.Gate, Check: req.Check.ID,
			Severity: domain.SeverityHigh, Message: "empty target URL",
		}}, nil
	}

	username := firstNonEmpty(req.Check.Config["username"], os.Getenv("LAB_WP_USER"), "admin")
	labPass := firstNonEmpty(req.Check.Config["password"], os.Getenv("LAB_WP_PASSWORD"))

	var findings []domain.Finding
	findings = append(findings, r.sprayLogin(ctx, req, base, username)...)
	findings = append(findings, r.xmlrpcMulticall(ctx, req, base, username)...)
	findings = append(findings, r.cookieFlags(ctx, req, base, username, labPass)...)
	findings = append(findings, r.hostHeaderReset(ctx, req, base)...)

	if len(findings) == 0 {
		findings = append(findings, f(req, "sec.auth.ok", domain.SeverityInfo, "auth abuse probes passed", base))
	}
	return findings, nil
}

func (r *Runner) sprayLogin(ctx context.Context, req ports.RunnerRequest, base, username string) []domain.Finding {
	// Intentionally fake passwords only — measure lockout, do not crack.
	passwords := []string{
		"lab-spray-never-1",
		"lab-spray-never-2",
		"lab-spray-never-3",
		"lab-spray-never-4",
		"lab-spray-never-5",
	}
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 12 * time.Second,
		Jar:     jar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	loginURL := base + "/wp-login.php"
	if _, err := client.Get(loginURL); err != nil {
		return []domain.Finding{f(req, "sec.auth.login_unreachable", domain.SeverityHigh, err.Error(), loginURL)}
	}

	locked := 0
	accepted := 0
	for _, pwd := range passwords {
		select {
		case <-ctx.Done():
			return []domain.Finding{f(req, "sec.auth.spray_canceled", domain.SeverityInfo, "spray canceled", loginURL)}
		default:
		}
		form := url.Values{}
		form.Set("log", username)
		form.Set("pwd", pwd)
		form.Set("wp-submit", "Log In")
		form.Set("redirect_to", base+"/wp-admin/")
		form.Set("testcookie", "1")
		hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
		if err != nil {
			continue
		}
		hreq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		resp, err := client.Do(hreq)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		_ = resp.Body.Close()
		low := strings.ToLower(string(body))
		loc := strings.ToLower(resp.Header.Get("Location"))

		if resp.StatusCode == 429 || looksLocked(low) {
			locked++
			continue
		}
		// Accidental success with spray password (should not happen)
		if resp.StatusCode == 302 && strings.Contains(loc, "wp-admin") && !strings.Contains(loc, "wp-login") {
			accepted++
			return []domain.Finding{f(req, "sec.auth.weak_password", domain.SeverityCritical,
				"login succeeded with spray wordlist password for user "+username, loginURL)}
		}
	}

	if locked > 0 {
		return []domain.Finding{f(req, "sec.auth.rate_limit_present", domain.SeverityInfo,
			"login rate limit / lockout signals observed during spray", loginURL)}
	}
	return []domain.Finding{f(req, "sec.auth.login_no_rate_limit", domain.SeverityHigh,
		"password spray ("+itoa(len(passwords))+" attempts) completed without lockout signal", loginURL)}
}

func (r *Runner) xmlrpcMulticall(ctx context.Context, req ports.RunnerRequest, base, username string) []domain.Finding {
	xmlrpc := base + "/xmlrpc.php"
	var calls strings.Builder
	calls.WriteString(`<?xml version="1.0"?><methodCall><methodName>system.multicall</methodName><params><param><value><array><data>`)
	for i := 1; i <= 8; i++ {
		calls.WriteString(`<value><struct>`)
		calls.WriteString(`<member><name>methodName</name><value><string>wp.getUsersBlogs</string></value></member>`)
		calls.WriteString(`<member><name>params</name><value><array><data>`)
		calls.WriteString(`<value><string>` + xmlEscape(username) + `</string></value>`)
		calls.WriteString(`<value><string>lab-xmlrpc-spray-` + itoa(i) + `</string></value>`)
		calls.WriteString(`</data></array></value></member></struct></value>`)
	}
	calls.WriteString(`</data></array></value></param></params></methodCall>`)

	client := &http.Client{Timeout: 15 * time.Second}
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, xmlrpc, strings.NewReader(calls.String()))
	if err != nil {
		return nil
	}
	hreq.Header.Set("Content-Type", "text/xml")
	resp, err := client.Do(hreq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 128*1024))
	s := string(body)
	if resp.StatusCode == 429 {
		return []domain.Finding{f(req, "sec.auth.xmlrpc_rate_limited", domain.SeverityInfo,
			"xmlrpc multicall received 429", xmlrpc)}
	}
	if resp.StatusCode == 200 && strings.Contains(s, "methodResponse") &&
		(strings.Contains(s, "faultCode") || strings.Contains(s, "isAdmin") || strings.Count(s, "<value>") >= 8) {
		return []domain.Finding{f(req, "sec.auth.xmlrpc_multicall", domain.SeverityHigh,
			"xmlrpc system.multicall accepted batched auth attempts without rate limit", xmlrpc)}
	}
	return nil
}

func (r *Runner) cookieFlags(ctx context.Context, req ports.RunnerRequest, base, username, labPass string) []domain.Finding {
	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Timeout: 12 * time.Second,
		Jar:     jar,
		// Do not follow redirects — Set-Cookie is on the 302 login response.
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	loginURL := base + "/wp-login.php"
	var findings []domain.Finding

	// Always inspect test cookie from login page (raw Set-Cookie for SameSite).
	resp, err := client.Get(loginURL)
	if err == nil {
		findings = append(findings, inspectSetCookieHeaders(req, resp.Header, loginURL, "wordpress_test_cookie", false)...)
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}

	if labPass == "" {
		findings = append(findings, f(req, "sec.auth.cookie_login_skipped", domain.SeverityInfo,
			"set LAB_WP_PASSWORD (or check config password) to inspect auth cookie flags", loginURL))
		return findings
	}

	form := url.Values{}
	form.Set("log", username)
	form.Set("pwd", labPass)
	form.Set("wp-submit", "Log In")
	form.Set("redirect_to", base+"/wp-admin/")
	form.Set("testcookie", "1")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, loginURL, strings.NewReader(form.Encode()))
	if err != nil {
		return findings
	}
	hreq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err = client.Do(hreq)
	if err != nil {
		return findings
	}
	rawHdr := resp.Header.Clone()
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()

	u, _ := url.Parse(base)
	loggedIn := false
	if u != nil {
		for _, c := range jar.Cookies(u) {
			if strings.HasPrefix(c.Name, "wordpress_logged_in_") {
				loggedIn = true
			}
		}
	}
	// Also detect from raw Set-Cookie if jar missed (redirect edge cases)
	for _, line := range rawHdr.Values("Set-Cookie") {
		if strings.HasPrefix(strings.ToLower(line), "wordpress_logged_in_") {
			loggedIn = true
		}
	}
	if !loggedIn {
		findings = append(findings, f(req, "sec.auth.login_failed", domain.SeverityMedium,
			"lab credentials did not produce wordpress_logged_in cookie", loginURL))
		return findings
	}
	findings = append(findings, inspectSetCookieHeaders(req, rawHdr, base, "wordpress_logged_in_", true)...)
	findings = append(findings, inspectSetCookieHeaders(req, rawHdr, base, "wordpress_sec_", true)...)
	return findings
}

// inspectSetCookieHeaders parses raw Set-Cookie (preserves SameSite; cookiejar may drop it).
func inspectSetCookieHeaders(req ports.RunnerRequest, hdr http.Header, rawURL, namePrefix string, authCookie bool) []domain.Finding {
	scheme := "http"
	if u, err := url.Parse(rawURL); err == nil && u.Scheme != "" {
		scheme = u.Scheme
	}
	var findings []domain.Finding
	seenSecureHTTP := false
	for _, line := range hdr.Values("Set-Cookie") {
		name := cookieName(line)
		if name == "" {
			continue
		}
		if name != namePrefix && !strings.HasPrefix(name, namePrefix) {
			continue
		}
		low := strings.ToLower(line)
		if !strings.Contains(low, "httponly") {
			findings = append(findings, f(req, "sec.auth.cookie_no_httponly", domain.SeverityMedium,
				"cookie "+name+" missing HttpOnly", rawURL))
		}
		if ss, ok := cookieAttr(low, "samesite"); !ok || ss == "" {
			findings = append(findings, f(req, "sec.auth.cookie_no_samesite", domain.SeverityMedium,
				"cookie "+name+" missing SameSite", rawURL))
		} else if ss == "none" && !strings.Contains(low, "secure") {
			findings = append(findings, f(req, "sec.auth.cookie_samesite_none_insecure", domain.SeverityHigh,
				"cookie "+name+" has SameSite=None without Secure", rawURL))
		}
		if authCookie && scheme == "https" && !strings.Contains(low, "secure") {
			findings = append(findings, f(req, "sec.auth.cookie_no_secure", domain.SeverityMedium,
				"cookie "+name+" missing Secure on HTTPS", rawURL))
		}
		if authCookie && scheme == "http" && !seenSecureHTTP {
			seenSecureHTTP = true
			findings = append(findings, f(req, "sec.auth.cookie_secure_http", domain.SeverityInfo,
				"auth cookies on plain HTTP — Secure flag not applicable until HTTPS", rawURL))
		}
	}
	return findings
}

func cookieName(setCookie string) string {
	part := strings.SplitN(setCookie, ";", 2)[0]
	kv := strings.SplitN(part, "=", 2)
	return strings.TrimSpace(kv[0])
}

func cookieAttr(lowLine, attr string) (string, bool) {
	// lowLine is already lowercased Set-Cookie
	parts := strings.Split(lowLine, ";")
	for _, p := range parts[1:] {
		p = strings.TrimSpace(p)
		if p == attr {
			return "", true
		}
		if strings.HasPrefix(p, attr+"=") {
			return strings.TrimSpace(strings.TrimPrefix(p, attr+"=")), true
		}
	}
	return "", false
}

func (r *Runner) hostHeaderReset(ctx context.Context, req ports.RunnerRequest, base string) []domain.Finding {
	resetURL := base + "/wp-login.php?action=lostpassword"
	body := strings.NewReader("user_login=admin&redirect_to=&wp-submit=Get+New+Password")
	hreq, err := http.NewRequestWithContext(ctx, http.MethodPost, resetURL, body)
	if err != nil {
		return nil
	}
	hreq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	hreq.Host = "evil.example"
	hreq.Header.Set("Host", "evil.example")
	client := &http.Client{
		Timeout: 12 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Do(hreq)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	low := strings.ToLower(string(b))
	if strings.Contains(low, "evil.example") {
		return []domain.Finding{f(req, "sec.auth.host_header_poison", domain.SeverityHigh,
			"password reset response reflects attacker Host evil.example", resetURL)}
	}
	// Soft note: probe executed; modern WP usually ignores Host for reset links.
	return []domain.Finding{f(req, "sec.auth.host_header_ok", domain.SeverityInfo,
		"Host-header poison probe did not reflect attacker host in response body", resetURL)}
}

func looksLocked(low string) bool {
	needles := []string{
		"too many", "limit login", "locked out", "locked", "try again later",
		"login attempts", "temporarily blocked", "rate limit",
	}
	for _, n := range needles {
		if strings.Contains(low, n) {
			return true
		}
	}
	return false
}

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [16]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func f(req ports.RunnerRequest, code string, sev domain.Severity, msg, target string) domain.Finding {
	return domain.Finding{Code: code, Gate: req.Gate, Check: req.Check.ID, Severity: sev, Message: msg, Target: target}
}
