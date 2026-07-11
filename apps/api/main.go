package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/fastygo/lab/packages/notify"
	"github.com/fastygo/lab/packages/presets"
	"github.com/fastygo/lab/packages/registry"
	"github.com/fastygo/lab/packages/runstore"
	runmem "github.com/fastygo/lab/packages/runstore/memory"
	runpg "github.com/fastygo/lab/packages/runstore/postgres"
	"github.com/fastygo/lab/packages/worker"
	"gopkg.in/yaml.v3"
)

// Cycle F API — runs + in-process worker + SSE + notify.
// Spec: .project/vps/cycle-f-saas.md
const version = "0.1.0-f4"

type server struct {
	store    runstore.Store
	backend  string
	repoRoot string
	queue    chan string
	wg       sync.WaitGroup
	notify   notify.Config
}

type createRunRequest struct {
	Preset       string            `json:"preset"`
	ManifestPath string            `json:"manifestPath"`
	Lab          string            `json:"lab"`
	Sync         bool              `json:"sync"`
	ThemeZip     string            `json:"themeZip"`
	Root         string            `json:"root"`
	BaseURL      string            `json:"baseUrl"`
	Config       map[string]string `json:"config"`
	CheckConfig  map[string]string `json:"checkConfig"`
}

func main() {
	addr := envOr("LAB_API_ADDR", ":8090")
	root := envOr("LAB_REPO_ROOT", registry.FindRepoRoot())
	store, backend, closer, err := openStore(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: store: %v\n", err)
		os.Exit(1)
	}
	if closer != nil {
		defer closer()
	}
	s := &server{
		store:    store,
		backend:  backend,
		repoRoot: root,
		queue:    make(chan string, 64),
		notify:   notify.FromEnv(),
	}
	workers := 1
	if n, err := strconv.Atoi(os.Getenv("LAB_API_WORKERS")); err == nil && n > 0 {
		workers = n
	}
	for i := 0; i < workers; i++ {
		s.wg.Add(1)
		go s.loop()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /v1/labs", s.handleLabs)
	mux.HandleFunc("GET /v1/presets", s.handlePresets)
	mux.HandleFunc("POST /v1/runs", s.handleCreateRun)
	mux.HandleFunc("GET /v1/runs", s.handleListRuns)
	mux.HandleFunc("GET /v1/runs/{id}", s.handleGetRun)
	mux.HandleFunc("GET /v1/runs/{id}/report", s.handleGetReport)
	mux.HandleFunc("GET /v1/runs/{id}/events", s.handleGetEvents)
	mux.HandleFunc("GET /v1/runs/{id}/events/stream", s.handleEventsStream)
	mux.HandleFunc("POST /v1/notify/test", s.handleNotifyTest)
	mux.HandleFunc("GET /v1/schedules", s.handleListSchedules)
	mux.HandleFunc("POST /v1/schedules", s.handleCreateSchedule)
	mux.HandleFunc("GET /v1/schedules/{id}", s.handleGetSchedule)
	mux.HandleFunc("PATCH /v1/schedules/{id}", s.handleUpdateSchedule)
	mux.HandleFunc("DELETE /v1/schedules/{id}", s.handleDeleteSchedule)
	mux.HandleFunc("POST /v1/schedules/{id}/fire", s.handleFireSchedule)

	go s.scheduleLoop()

	fmt.Fprintf(os.Stderr, "lab-api %s listening on %s (repo=%s workers=%d store=%s notify=%v)\n", version, addr, root, workers, backend, s.notify.Enabled())
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// openStore picks Postgres when a connection URL is set (Supabase-style DATABASE_URL).
func openStore(ctx context.Context) (runstore.Store, string, func(), error) {
	url := os.Getenv("LAB_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		return runmem.New(), "memory", nil, nil
	}
	pg, err := runpg.Open(ctx, url)
	if err != nil {
		return nil, "", nil, err
	}
	return pg, "postgres", pg.Close, nil
}

func (s *server) loop() {
	defer s.wg.Done()
	for id := range s.queue {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
		err := worker.Execute(ctx, worker.Options{Store: s.store, RepoRoot: s.repoRoot}, id)
		cancel()
		if err != nil {
			fmt.Fprintf(os.Stderr, "worker run %s: %v\n", id, err)
		}
		s.afterRun(id)
	}
}

func (s *server) afterRun(id string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	run, err := s.store.GetRun(ctx, id)
	if err != nil {
		return
	}
	if err := notify.NotifyRunFinished(ctx, s.notify, run); err != nil {
		fmt.Fprintf(os.Stderr, "notify run %s: %v\n", id, err)
	}
}

func (s *server) handleNotifyTest(w http.ResponseWriter, r *http.Request) {
	if !s.notify.Enabled() {
		writeErr(w, http.StatusServiceUnavailable, "notify not configured (SLACK_WEBHOOK_URL and/or TELEGRAM_BOT_TOKEN+TELEGRAM_CHAT_ID)")
		return
	}
	demo := &runstore.Run{
		ID:     "notify-test",
		Lab:    "demo",
		Status: runstore.StatusFail,
		Report: nil,
	}
	// Force send regardless of filter for test endpoint.
	cfg := s.notify
	cfg.Filter = notify.FilterAlways
	if err := notify.NotifyRunFinished(r.Context(), cfg, demo); err != nil {
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "message": notify.Message(demo, cfg.DashboardBase)})
}

func (s *server) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"service": "lab-api",
		"version": version,
		"cycle":   "F4",
		"store":   s.backend,
		"notify":  s.notify.Enabled(),
		"ts":      time.Now().UTC().Format(time.RFC3339),
		"roadmap": "see .project/vps/cycle-f-saas.md",
	})
}

func (s *server) handleLabs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"labs": registry.KnownLabs()})
}

func (s *server) handlePresets(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"presets": presets.List()})
}

func (s *server) handleCreateRun(w http.ResponseWriter, r *http.Request) {
	var req createRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	key := req.Preset
	if key == "" {
		key = req.ManifestPath
	}
	if key == "" {
		writeErr(w, http.StatusBadRequest, "preset or manifestPath required")
		return
	}
	m, path, err := presets.Load(s.repoRoot, key)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	presets.ApplyBindings(m, presets.Bindings{
		ThemeZip:    req.ThemeZip,
		Root:        req.Root,
		BaseURL:     req.BaseURL,
		Config:      req.Config,
		CheckConfig: req.CheckConfig,
	})
	raw, err := yaml.Marshal(m)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "marshal manifest")
		return
	}
	lab := m.Spec.Lab
	if req.Lab != "" {
		lab = req.Lab
	}
	now := time.Now().UTC()
	run := &runstore.Run{
		Lab:          lab,
		Status:       runstore.StatusQueued,
		ManifestJSON: raw,
		CreatedAt:    now,
	}
	if err := s.store.CreateRun(r.Context(), run); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	sync := req.Sync || r.URL.Query().Get("sync") == "1"
	if sync {
		ctx, cancel := context.WithTimeout(r.Context(), 60*time.Minute)
		defer cancel()
		_ = worker.Execute(ctx, worker.Options{Store: s.store, RepoRoot: s.repoRoot}, run.ID)
		s.afterRun(run.ID)
		updated, err := s.store.GetRun(r.Context(), run.ID)
		if err != nil {
			writeErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, runView(updated, path))
		return
	}
	select {
	case s.queue <- run.ID:
	default:
		writeErr(w, http.StatusServiceUnavailable, "worker queue full")
		return
	}
	writeJSON(w, http.StatusAccepted, runView(run, path))
}

func (s *server) handleListRuns(w http.ResponseWriter, r *http.Request) {
	lab := r.URL.Query().Get("lab")
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	runs, err := s.store.ListRuns(r.Context(), lab, limit)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(runs))
	for _, run := range runs {
		out = append(out, runView(run, ""))
	}
	writeJSON(w, http.StatusOK, map[string]any{"runs": out})
}

func (s *server) handleGetRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	events, _ := s.store.ListEvents(r.Context(), id)
	view := runView(run, "")
	view["eventCount"] = len(events)
	if len(events) > 0 {
		view["lastEvent"] = events[len(events)-1].Type
	}
	writeJSON(w, http.StatusOK, view)
}

func (s *server) handleGetReport(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	if run.Report == nil {
		writeErr(w, http.StatusConflict, "report not ready")
		return
	}
	writeJSON(w, http.StatusOK, run.Report)
}

func (s *server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	events, err := s.store.ListEvents(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"runId": id, "events": events})
}

func runView(run *runstore.Run, manifestPath string) map[string]any {
	v := map[string]any{
		"id":        run.ID,
		"lab":       run.Lab,
		"status":    run.Status,
		"createdAt": run.CreatedAt,
		"error":     run.Error,
	}
	if manifestPath != "" {
		v["manifestPath"] = filepath.ToSlash(manifestPath)
	}
	if run.StartedAt != nil {
		v["startedAt"] = run.StartedAt
	}
	if run.FinishedAt != nil {
		v["finishedAt"] = run.FinishedAt
	}
	if run.Report != nil {
		v["reportStatus"] = run.Report.Status
		v["summary"] = run.Report.Summary
	}
	return v
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
