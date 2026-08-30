/*
package themesec_test

import (
	"archive/zip"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
	"github.com/fastygo/lab/packages/orchestrator/themesec"
)

func TestThemeSecStaticDanger(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bad.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	w, err := zw.Create("bad/functions.php")
	if err != nil {
		t.Fatal(err)
	}
	danger := "<?php\n" + "ev" + "al" + "(" + "$_" + "GET" + "['x']);\n"
	_, _ = w.Write([]byte(danger))
	_ = zw.Close()
	_ = zf.Close()

	r := themesec.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "S4",
		Check: domain.Check{ID: "theme-sec", Runner: "theme-sec"},
		Target: domain.Target{
			BaseURL:  "",
			Metadata: map[string]string{"themeZip": zipPath},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Code == "sec.theme.eval" {
			return
		}
	}
	t.Fatalf("expected sec.theme.eval, got %+v", got)
}

func TestThemeSecXSSReflected(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("s")
		_, _ = w.Write([]byte("<html><title>Search: " + q + "</title></html>"))
	}))
	t.Cleanup(srv.Close)

	r := themesec.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S4",
		Check:  domain.Check{ID: "theme-sec", Runner: "theme-sec"},
		Target: domain.Target{BaseURL: srv.URL},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Code == "sec.theme.xss_reflected" {
			return
		}
	}
	t.Fatalf("expected xss_reflected, got %+v", got)
}

func TestThemeSecXSSOK(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html><title>Search: &lt;script&gt;alert(1)&lt;/script&gt;</title></html>`))
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	zipPath := filepath.Join(dir, "ok.zip")
	zf, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	w, err := zw.Create("ok/functions.php")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = w.Write([]byte("<?php\necho 'hi';\n"))
	_ = zw.Close()
	_ = zf.Close()

	r := themesec.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "S4",
		Check: domain.Check{ID: "theme-sec", Runner: "theme-sec"},
		Target: domain.Target{
			BaseURL:  srv.URL,
			Metadata: map[string]string{"themeZip": zipPath},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Code == "sec.theme.xss_reflected" || f.Code == "sec.theme.eval" {
			t.Fatalf("unexpected finding %+v", f)
		}
	}
}
*/