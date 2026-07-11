package seometa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

func TestSEOMetaHappyPath(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head>
<title>Hello</title>
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="description" content="desc">
</head><body><h1>Hi</h1></body></html>`))
	}))
	defer srv.Close()

	r := New()
	findings, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "Q5-seo",
		Check: domain.Check{ID: "seo", Runner: "seo-meta"},
		Target: domain.Target{BaseURL: srv.URL},
		URLs:   []string{srv.URL + "/"},
	})
	if err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, f := range findings {
		codes[f.Code] = true
	}
	for _, want := range []string{"quality.seo.ok", "quality.seo.title_ok", "quality.seo.viewport_ok", "quality.seo.h1_ok", "quality.seo.summary"} {
		if !codes[want] {
			t.Fatalf("missing %s in %#v", want, codes)
		}
	}
}

func TestSEOMetaMissingTitle(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head>
<meta name="viewport" content="width=device-width">
</head><body><h1>X</h1></body></html>`))
	}))
	defer srv.Close()

	r := New()
	findings, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "Q5",
		Check: domain.Check{ID: "seo"},
		Target: domain.Target{BaseURL: srv.URL},
		URLs:   []string{srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range findings {
		if f.Code == "quality.seo.title_missing" {
			found = true
		}
	}
	if !found {
		t.Fatalf("%+v", findings)
	}
}
