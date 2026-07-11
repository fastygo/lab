package worker_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/fastygo/lab/packages/presets"
	"github.com/fastygo/lab/packages/runstore"
	runmem "github.com/fastygo/lab/packages/runstore/memory"
	"github.com/fastygo/lab/packages/worker"
	"gopkg.in/yaml.v3"
)

func TestWorkerExecuteDemo(t *testing.T) {
	t.Parallel()
	root := findRepoRoot(t)
	m, _, err := presets.Load(root, "demo")
	if err != nil {
		t.Fatal(err)
	}
	raw, err := yaml.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	store := runmem.New()
	run := &runstore.Run{
		Lab:          "demo",
		Status:       runstore.StatusQueued,
		ManifestJSON: raw,
		CreatedAt:    time.Now().UTC(),
	}
	if err := store.CreateRun(context.Background(), run); err != nil {
		t.Fatal(err)
	}
	if err := worker.Execute(context.Background(), worker.Options{Store: store, RepoRoot: root}, run.ID); err != nil {
		t.Fatal(err)
	}
	got, err := store.GetRun(context.Background(), run.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != runstore.StatusPass {
		t.Fatalf("status=%s err=%s", got.Status, got.Error)
	}
	if got.Report == nil || got.Report.Summary.Total < 1 {
		t.Fatalf("report=%v", got.Report)
	}
	events, err := store.ListEvents(context.Background(), run.ID)
	if err != nil || len(events) < 5 {
		t.Fatalf("events=%d err=%v", len(events), err)
	}
	if events[0].Type != "run.started" {
		t.Fatalf("first=%s", events[0].Type)
	}
	if events[len(events)-1].Type != "run.finished" {
		t.Fatalf("last=%s", events[len(events)-1].Type)
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
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
