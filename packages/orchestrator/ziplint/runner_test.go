package ziplint_test

import (
	"archive/zip"
	"bytes"
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
	"github.com/fastygo/lab/packages/orchestrator/ziplint"
)

func TestZipLintOK(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "good.zip")
	writeThemeZip(t, zipPath, true, false)
	r := ziplint.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "C1",
		Check: domain.Check{ID: "zip-lint", Runner: "zip-lint"},
		Target: domain.Target{Metadata: map[string]string{"themeZip": zipPath}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range got {
		if f.Severity == domain.SeverityHigh || f.Severity == domain.SeverityCritical {
			t.Fatalf("unexpected high finding: %+v", f)
		}
	}
}

func TestZipLintMissingAndA11yTag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bad.zip")
	writeThemeZip(t, zipPath, false, true)
	r := ziplint.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "C1",
		Check: domain.Check{ID: "zip-lint", Runner: "zip-lint"},
		Target: domain.Target{Metadata: map[string]string{"themeZip": zipPath}},
	})
	if err != nil {
		t.Fatal(err)
	}
	codes := map[string]bool{}
	for _, f := range got {
		codes[f.Code] = true
	}
	if !codes["org.zip.missing_readme_txt"] {
		t.Fatalf("expected missing readme, got %+v", got)
	}
	if !codes["org.zip.tag_accessibility_ready"] {
		t.Fatalf("expected a11y tag block, got %+v", got)
	}
}

func writeThemeZip(t *testing.T, path string, withReadme, a11yTag bool) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	style := "/*\nTheme Name: Fixture\nVersion: 1.0.0\nText Domain: fixture\n"
	if a11yTag {
		style += "Tags: blog, accessibility-ready\n"
	} else {
		style += "Tags: blog\n"
	}
	style += "*/\n"
	mustZip(t, zw, "fixture/style.css", []byte(style))
	mustZip(t, zw, "fixture/functions.php", []byte("<?php\n"))
	mustZip(t, zw, "fixture/LICENSE", []byte("MIT\n"))
	if withReadme {
		mustZip(t, zw, "fixture/readme.txt", []byte("=== Fixture ===\n"))
	}
	mustZip(t, zw, "fixture/screenshot.png", tinyPNG(t, 800, 600))
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}

func mustZip(t *testing.T, zw *zip.Writer, name string, data []byte) {
	t.Helper()
	w, err := zw.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
}

func tinyPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}
