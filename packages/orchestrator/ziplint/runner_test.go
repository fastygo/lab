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
	writeThemeZip(t, zipPath, themeOpts{
		withReadme: true,
		resources:  true,
		assetJS:    true,
		withMinTwin: true,
	})
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
			t.Fatalf("unexpected high finding: %+v\nall=%+v", f, got)
		}
	}
}

func TestZipLintMissingAndA11yTag(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "bad.zip")
	writeThemeZip(t, zipPath, themeOpts{a11yTag: true})
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

func TestZipLintForbiddenExtAndPolicy(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "policy.zip")
	writeThemeZip(t, zipPath, themeOpts{
		withReadme: true,
		resources:  true,
		badXML:     true,
		badSH:      true,
		nestedZip:  true,
		cpt:        true,
		shortcode:  true,
		wrongDomain: true,
	})
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
	for _, want := range []string{
		"org.zip.forbidden_ext_xml",
		"org.zip.forbidden_ext_sh",
		"org.zip.nested_zip",
		"org.zip.policy_cpt",
		"org.zip.policy_shortcode",
		"org.zip.style_textdomain_slug",
	} {
		if !codes[want] {
			t.Fatalf("missing %s in %+v", want, got)
		}
	}
}

func TestZipLintMinifiedWithoutSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "min.zip")
	writeThemeZip(t, zipPath, themeOpts{
		withReadme:  true,
		resources:   true,
		assetJS:     true,
		withMinTwin: false, // only .min.js
		onlyMin:     true,
	})
	r := ziplint.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "C1",
		Check: domain.Check{ID: "zip-lint", Runner: "zip-lint"},
		Target: domain.Target{Metadata: map[string]string{"themeZip": zipPath}},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range got {
		if f.Code == "org.zip.minified_without_source" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected minified_without_source, got %+v", got)
	}
}

func TestZipLintResourcesUnattributed(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "res.zip")
	writeThemeZip(t, zipPath, themeOpts{
		withReadme:  true,
		resources:   false, // readme without Resources but has assets
		assetJS:     true,
		withMinTwin: true,
	})
	r := ziplint.New()
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "C1",
		Check: domain.Check{ID: "zip-lint", Runner: "zip-lint"},
		Target: domain.Target{Metadata: map[string]string{"themeZip": zipPath}},
	})
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, f := range got {
		if f.Code == "org.zip.resources_missing_section" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected resources_missing_section, got %+v", got)
	}
}

type themeOpts struct {
	withReadme  bool
	a11yTag     bool
	resources   bool
	assetJS     bool
	withMinTwin bool
	onlyMin     bool
	badXML      bool
	badSH       bool
	nestedZip   bool
	cpt         bool
	shortcode   bool
	wrongDomain bool
}

func writeThemeZip(t *testing.T, path string, o themeOpts) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)

	domain := "fixture"
	if o.wrongDomain {
		domain = "not-fixture"
	}
	style := "/*\nTheme Name: Fixture\nVersion: 1.0.0\nRequires at least: 6.4\nTested up to: 6.7\nRequires PHP: 8.2\nText Domain: " + domain + "\n"
	if o.a11yTag {
		style += "Tags: blog, accessibility-ready\n"
	} else {
		style += "Tags: blog\n"
	}
	style += "*/\n"
	mustZip(t, zw, "fixture/style.css", []byte(style))

	php := "<?php\n"
	if o.cpt {
		php += "register_post_type('book', []);\n"
	}
	if o.shortcode {
		php += "add_shortcode('x', 'y');\n"
	}
	mustZip(t, zw, "fixture/functions.php", []byte(php))
	mustZip(t, zw, "fixture/LICENSE", []byte("MIT\n"))
	if o.withReadme {
		readme := "=== Fixture ===\n\n"
		if o.resources {
			readme += "== Resources ==\n* ui8kit.js — MIT\n* theme — theme author\n"
		}
		mustZip(t, zw, "fixture/readme.txt", []byte(readme))
	}
	mustZip(t, zw, "fixture/screenshot.png", tinyPNG(t, 800, 600))

	if o.assetJS || o.onlyMin {
		if o.onlyMin || o.withMinTwin {
			mustZip(t, zw, "fixture/assets/js/ui8kit.min.js", []byte("/*!min*/\n"))
		}
		if o.withMinTwin && !o.onlyMin {
			mustZip(t, zw, "fixture/assets/js/ui8kit.js", []byte("// source\n"))
		}
		if o.assetJS && !o.onlyMin && !o.withMinTwin {
			mustZip(t, zw, "fixture/assets/js/ui8kit.js", []byte("// source\n"))
		}
	}
	if o.badXML {
		mustZip(t, zw, "fixture/data/export.xml", []byte("<a/>\n"))
	}
	if o.badSH {
		mustZip(t, zw, "fixture/bin/setup.sh", []byte("#!/bin/sh\n"))
	}
	if o.nestedZip {
		mustZip(t, zw, "fixture/extra/theme.zip", []byte("PK\x03\x04fake"))
	}

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
