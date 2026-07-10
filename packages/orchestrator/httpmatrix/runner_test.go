package httpmatrix_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/httpmatrix"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

func TestHTTPSmokeOKAnd404(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("home"))
	})
	mux.HandleFunc("/about/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	mux.HandleFunc("/this-page-does-not-exist-lab-404/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	r := httpmatrix.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "C3",
		Check: domain.Check{ID: "url-matrix", Runner: "http-matrix"},
		Target: domain.Target{BaseURL: srv.URL},
		URLs: []string{
			srv.URL + "/",
			srv.URL + "/about/",
			srv.URL + "/this-page-does-not-exist-lab-404/",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	var summary string
	for _, f := range got {
		if f.Severity == domain.SeverityHigh || f.Severity == domain.SeverityCritical {
			t.Fatalf("unexpected high: %+v", f)
		}
		if f.Code == "org.matrix.smoke_summary" {
			summary = f.Message
		}
	}
	if !strings.Contains(summary, "3/3") {
		t.Fatalf("summary=%q findings=%+v", summary, got)
	}
}

func TestHTTPSmoke5xx(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	t.Cleanup(srv.Close)

	r := httpmatrix.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "C3",
		Check:  domain.Check{ID: "m", Runner: "http-matrix"},
		Target: domain.Target{BaseURL: srv.URL},
		URLs:   []string{srv.URL + "/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range got {
		if f.Code == "org.matrix.status_5xx" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%+v", got)
	}
}
