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
