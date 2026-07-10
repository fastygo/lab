package domain_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/fastygo/lab/packages/domain"
)

func TestLoadDemoManifest(t *testing.T) {
	t.Parallel()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("caller")
	}
	// packages/domain -> repo root
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	path := filepath.Join(root, "testdata", "manifests", "demo.lab.yaml")
	m, err := domain.LoadManifest(path)
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.Lab != "demo" {
		t.Fatalf("lab=%s", m.Spec.Lab)
	}
	if m.Spec.Adapter.ID != "noop" {
		t.Fatalf("adapter=%s", m.Spec.Adapter.ID)
	}
	if len(m.Spec.Gates) != 1 || len(m.Spec.Gates[0].Checks) != 2 {
		t.Fatalf("gates=%+v", m.Spec.Gates)
	}
}
