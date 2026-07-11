package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/fastygo/lab/packages/registry"
	runmem "github.com/fastygo/lab/packages/runstore/memory"
)

func TestCreateRunDemoSync(t *testing.T) {
	root := registry.FindRepoRoot()
	if root == "" {
		wd, _ := os.Getwd()
		root = findGoMod(wd)
	}
	s := &server{
		store:    runmem.New(),
		repoRoot: root,
		queue:    make(chan string, 8),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/runs", s.handleCreateRun)
	mux.HandleFunc("GET /v1/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /v1/runs/{id}/report", s.handleGetReport)
	mux.HandleFunc("GET /v1/runs/{id}/events", s.handleGetEvents)

	body := []byte(`{"preset":"demo","sync":true}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("missing id")
	}
	if created["status"] != "pass" {
		t.Fatalf("status=%v", created["status"])
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/runs/"+id+"/report", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("report status=%d body=%s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/v1/runs/"+id+"/events", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("events status=%d", rec.Code)
	}
	var evWrap map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &evWrap)
	events, _ := evWrap["events"].([]any)
	if len(events) < 5 {
		t.Fatalf("events=%d", len(events))
	}
}

func TestCreateRunBindingsInSnapshot(t *testing.T) {
	root := registry.FindRepoRoot()
	if root == "" {
		wd, _ := os.Getwd()
		root = findGoMod(wd)
	}
	store := runmem.New()
	s := &server{store: store, repoRoot: root, queue: make(chan string, 8)}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/runs", s.handleCreateRun)

	// Async so we don't need Docker for org/quality — just assert queued snapshot.
	body := []byte(`{
		"preset":"org",
		"themeZip":"testdata/dist/custom.zip",
		"baseUrl":"http://127.0.0.1:9999",
		"checkConfig":{"dockerNetwork":"labnet"}
	}`)
	req := httptest.NewRequest(http.MethodPost, "/v1/runs", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id, _ := created["id"].(string)
	run, err := store.GetRun(context.Background(), id)
	if err != nil {
		t.Fatal(err)
	}
	snap := string(run.ManifestJSON)
	if !bytes.Contains(run.ManifestJSON, []byte("testdata/dist/custom.zip")) {
		t.Fatalf("themeZip not in snapshot: %s", snap)
	}
	if !bytes.Contains(run.ManifestJSON, []byte("http://127.0.0.1:9999")) {
		t.Fatalf("baseUrl not in snapshot")
	}
	if !bytes.Contains(run.ManifestJSON, []byte("labnet")) {
		t.Fatalf("checkConfig not in snapshot")
	}
	// Drain queue so worker does not race after test (do not execute).
	select {
	case <-s.queue:
	default:
	}
}

func findGoMod(dir string) string {
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
