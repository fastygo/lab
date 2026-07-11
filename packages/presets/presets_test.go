package presets_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/fastygo/lab/packages/presets"
)

func TestLoadDemo(t *testing.T) {
	t.Parallel()
	root := findRepoRoot(t)
	m, path, err := presets.Load(root, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.Lab != "demo" {
		t.Fatalf("lab=%s", m.Spec.Lab)
	}
	if path == "" {
		t.Fatal("empty path")
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, _ := os.Getwd()
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found")
		}
		dir = parent
	}
}
