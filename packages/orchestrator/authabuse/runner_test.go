package authabuse_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/authabuse"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

func TestAuthAbuseNoRateLimitAndMulticall(t *testing.T) {
	t.Parallel()
	loginHits := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/wp-login.php" && r.Method == http.MethodGet:
			w.Header().Set("Set-Cookie", "wordpress_test_cookie=1; path=/")
			w.WriteHeader(200)
			_, _ = w.Write([]byte("login"))
		case r.URL.Path == "/wp-login.php" && r.Method == http.MethodPost && r.URL.Query().Get("action") == "lostpassword":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("check your email"))
		case r.URL.Path == "/wp-login.php" && r.Method == http.MethodPost:
			loginHits++
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`<div id="login_error">Error: incorrect password</div>`))
		case r.URL.Path == "/xmlrpc.php":
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`<?xml version="1.0"?><methodResponse><params><param><value><array><data>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
<value><struct><member><name>faultCode</name><value><int>403</int></value></member></struct></value>
</data></array></value></param></params></methodResponse>`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	r := authabuse.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S3",
		Check:  domain.Check{ID: "auth", Runner: "auth-abuse"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, f := range got {
		codes[f.Code] = true
	}
	if !codes["sec.auth.login_no_rate_limit"] {
		t.Fatalf("expected login_no_rate_limit, got %+v", got)
	}
	if !codes["sec.auth.xmlrpc_multicall"] {
		t.Fatalf("expected xmlrpc_multicall, got %+v", got)
	}
	if !codes["sec.auth.cookie_login_skipped"] {
		t.Fatalf("expected cookie_login_skipped, got %+v", got)
	}
	if loginHits < 5 {
		t.Fatalf("expected >=5 spray posts, got %d", loginHits)
	}
}

func TestAuthAbuseRateLimit(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/xmlrpc.php" {
			w.WriteHeader(405)
			return
		}
		if r.Method == http.MethodPost && r.URL.Query().Get("action") != "lostpassword" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ERROR: Too many failed login attempts."))
			return
		}
		if r.URL.Query().Get("action") == "lostpassword" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte("login"))
	}))
	t.Cleanup(srv.Close)

	r := authabuse.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S3",
		Check:  domain.Check{ID: "auth", Runner: "auth-abuse"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Code == "sec.auth.rate_limit_present" {
			return
		}
	}
	t.Fatalf("expected rate_limit_present, got %+v", got)
}

func TestAuthAbuseHostPoison(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("action") == "lostpassword" && r.Method == http.MethodPost {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("Reset link: http://evil.example/wp-login.php?action=rp"))
			return
		}
		if r.URL.Path == "/xmlrpc.php" {
			w.WriteHeader(404)
			return
		}
		if r.Method == http.MethodPost {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ERROR: Too many failed login attempts."))
			return
		}
		w.WriteHeader(200)
	}))
	t.Cleanup(srv.Close)

	r := authabuse.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S3",
		Check:  domain.Check{ID: "auth", Runner: "auth-abuse"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Code == "sec.auth.host_header_poison" {
			if !strings.Contains(f.Message, "evil.example") {
				t.Fatalf("message: %s", f.Message)
			}
			return
		}
	}
	t.Fatalf("expected host_header_poison, got %+v", got)
}

func TestAuthAbuseSameSite(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/wp-login.php" && r.Method == http.MethodGet:
			w.Header().Add("Set-Cookie", "wordpress_test_cookie=1; path=/; HttpOnly; SameSite=Lax")
			w.WriteHeader(200)
		case r.URL.Path == "/wp-login.php" && r.Method == http.MethodPost && r.URL.Query().Get("action") == "lostpassword":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
		case r.URL.Path == "/wp-login.php" && r.Method == http.MethodPost:
			_ = r.ParseForm()
			if r.Form.Get("pwd") == "lab-pass" {
				w.Header().Add("Set-Cookie", "wordpress_logged_in_abc=session; path=/; HttpOnly")
				w.Header().Set("Location", "/wp-admin/")
				w.WriteHeader(302)
				return
			}
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`<div id="login_error">Error: incorrect password</div>`))
		case r.URL.Path == "/xmlrpc.php":
			w.WriteHeader(405)
		default:
			w.WriteHeader(200)
		}
	}))
	t.Cleanup(srv.Close)

	r := authabuse.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S3",
		Check:  domain.Check{ID: "auth", Runner: "auth-abuse", Config: map[string]string{"password": "lab-pass"}},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, f := range got {
		codes[f.Code] = true
	}
	if !codes["sec.auth.cookie_no_samesite"] {
		t.Fatalf("expected cookie_no_samesite for logged_in cookie, got %+v", got)
	}
	if codes["sec.auth.weak_password"] {
		t.Fatalf("spray must not succeed: %+v", got)
	}
}
