package wordpress_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastygo/lab/packages/adapters/wordpress"
)

func TestPrepareResolvesRelativeThemeZip(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zipPath := filepath.Join(dir, "latte.zip")
	if err := os.WriteFile(zipPath, []byte("PK"), 0o644); err != nil {
		t.Fatal(err)
	}
	a := wordpress.New(dir)
	if err := a.Prepare(context.Background(), map[string]string{
		"baseUrl":  "http://127.0.0.1:8080",
		"themeZip": "latte.zip",
	}); err != nil {
		t.Fatal(err)
	}
	target, err := a.Serve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if target.Metadata["themeZip"] != zipPath {
		t.Fatalf("got %q want %q", target.Metadata["themeZip"], zipPath)
	}
	urls, err := a.Matrix(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(urls) < 5 {
		t.Fatalf("matrix too small: %v", urls)
	}
}

func TestAllowedBaseURL(t *testing.T) {
	a := wordpress.New(t.TempDir())
	t.Setenv("LAB_ALLOWED_BASE_URLS", "http://127.0.0.1:8080,http://5.129.242.217:8080")
	if err := a.Prepare(context.Background(), map[string]string{"baseUrl": "http://evil.example"}); err == nil {
		t.Fatal("expected deny")
	}
	if err := a.Prepare(context.Background(), map[string]string{"baseUrl": "http://127.0.0.1:8080"}); err != nil {
		t.Fatal(err)
	}
}

func TestMatrixIncludesAttachmentWhenSeeded(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	seed := filepath.Join(dir, "org-seed.json")
	if err := os.WriteFile(seed, []byte(`{"attachmentId":"42","postId":"7","pageId":"9","catId":"3","tagSlug":"markup","imported":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	a := wordpress.New(dir)
	if err := a.Prepare(context.Background(), map[string]string{
		"baseUrl":  "http://127.0.0.1:8080",
		"seedFile": seed,
	}); err != nil {
		t.Fatal(err)
	}
	urls, err := a.Matrix(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	wantAttach := "http://127.0.0.1:8080/?attachment_id=42"
	wantPost := "http://127.0.0.1:8080/?p=7"
	var hasAttach, hasPost bool
	for _, u := range urls {
		if u == wantAttach {
			hasAttach = true
		}
		if u == wantPost {
			hasPost = true
		}
	}
	if !hasAttach {
		t.Fatalf("missing attachment URL in %v", urls)
	}
	if !hasPost {
		t.Fatalf("missing seeded post URL in %v", urls)
	}
}
