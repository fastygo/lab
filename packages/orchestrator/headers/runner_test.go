package headers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/headers"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

func TestHeadersMissing(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(srv.Close)

	r := headers.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S5",
		Check:  domain.Check{ID: "headers", Runner: "headers"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, f := range got {
		codes[f.Code] = true
	}
	for _, want := range []string{"sec.headers.nosniff", "sec.headers.csp", "sec.headers.hsts", "sec.headers.permissions"} {
		if !codes[want] {
			t.Fatalf("expected %s: %+v", want, got)
		}
	}
}

func TestHeadersOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Strict-Transport-Security", "max-age=31536000")
			w.Header().Set("Permissions-Policy", "camera=()")
			w.WriteHeader(200)
			_, _ = w.Write([]byte("<html><body>ok</body></html>"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	r := headers.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S5",
		Check:  domain.Check{ID: "headers", Runner: "headers"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if strings.HasPrefix(f.Code, "sec.headers.") && f.Code != "sec.headers.ok" {
			t.Fatalf("unexpected header finding: %+v", f)
		}
	}
}

func TestReconFindings(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/" && r.URL.Query().Get("author") == "1":
			http.Redirect(w, r, "/author/admin/", http.StatusFound)
		case r.URL.Path == "/":
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("Referrer-Policy", "no-referrer")
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
			w.Header().Set("Strict-Transport-Security", "max-age=1")
			w.Header().Set("Permissions-Policy", "camera=()")
			_, _ = w.Write([]byte(`<meta name="generator" content="WordPress 6.7.1" />`))
		case r.URL.Path == "/readme.html":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("WordPress"))
		case r.URL.Path == "/xmlrpc.php" && r.Method == http.MethodPost:
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`<?xml version="1.0"?><methodResponse><params></params></methodResponse>`))
		case r.URL.Path == "/author/admin/" || r.URL.Path == "/author/admin":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("author"))
		case r.URL.Path == "/wp-json/wp/v2/users":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"id":1,"slug":"admin","name":"Admin"}]`))
		case r.URL.Path == "/wp-json/" || r.URL.Path == "/wp-json":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"namespaces":["wp/v2"],"routes":{}}`))
		case r.URL.Path == "/wp-login.php" && r.URL.Query().Get("action") == "register":
			_, _ = w.Write([]byte(`<form id="registerform"><input name="user_login"/><input name="user_email"/></form>`))
		case r.URL.Path == "/wp-content/uploads/" || r.URL.Path == "/wp-content/uploads":
			_, _ = w.Write([]byte(`<title>Index of /wp-content/uploads</title>`))
		case r.URL.Path == "/.env":
			_, _ = w.Write([]byte("DB_PASSWORD=secret"))
		case r.URL.Path == "/wp-cron.php":
			w.WriteHeader(200)
		case r.URL.Path == "/wp-admin/theme-editor.php":
			http.Redirect(w, r, "/wp-login.php", http.StatusFound)
		case r.URL.Path == "/wp-login.php":
			w.WriteHeader(200)
			_, _ = w.Write([]byte("login"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	r := headers.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S1",
		Check:  domain.Check{ID: "recon", Runner: "headers"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, f := range got {
		codes[f.Code] = true
	}
	want := []string{
		"sec.recon.generator",
		"sec.recon.readme",
		"sec.recon.xmlrpc",
		"sec.recon.user_enum.author",
		"sec.recon.user_enum.rest",
		"sec.recon.rest_index",
		"sec.recon.registration",
		"sec.recon.dir_listing.uploads",
		"sec.recon.sensitive.env",
		"sec.recon.wp_cron",
		"sec.config.file_edit",
	}
	for _, c := range want {
		if !codes[c] {
			t.Fatalf("missing %s in %+v", c, got)
		}
	}
}
