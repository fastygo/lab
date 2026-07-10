package headers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
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
	if !codes["sec.headers.nosniff"] {
		t.Fatalf("expected nosniff finding: %+v", got)
	}
}

func TestHeadersOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.WriteHeader(200)
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
		if f.Code == "sec.headers.nosniff" {
			t.Fatalf("unexpected nosniff: %+v", got)
		}
	}
}
