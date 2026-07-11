package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/fastygo/lab/packages/presets"
	"github.com/fastygo/lab/packages/runstore"
	"github.com/fastygo/lab/packages/scheduler"
	"gopkg.in/yaml.v3"
)

type createScheduleRequest struct {
	Cron     string `json:"cron"`
	Preset   string `json:"preset"`
	Lab      string `json:"lab"`
	Enabled  *bool  `json:"enabled"`
	ThemeZip string `json:"themeZip"`
	BaseURL  string `json:"baseUrl"`
	Root     string `json:"root"`
}

func (s *server) scheduleLoop() {
	interval := 30 * time.Second
	if v := os.Getenv("LAB_SCHEDULER_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d >= time.Second {
			interval = d
		}
	}
	t := time.NewTicker(interval)
	defer t.Stop()
	s.tickSchedules()
	for range t.C {
		s.tickSchedules()
	}
}

func (s *server) tickSchedules() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	list, err := s.store.ListSchedules(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "scheduler list: %v\n", err)
		return
	}
	now := time.Now().UTC()
	for _, sch := range list {
		if !sch.Enabled {
			continue
		}
		due, next, err := scheduler.IsDue(sch.Cron, sch.LastRunAt, sch.NextRunAt, now)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scheduler %s cron: %v\n", sch.ID, err)
			continue
		}
		if !due {
			if sch.NextRunAt == nil || !sch.NextRunAt.Equal(next) {
				sch.NextRunAt = &next
				_ = s.store.UpdateSchedule(ctx, sch)
			}
			continue
		}
		runID, err := s.enqueuePreset(ctx, sch.Preset, sch.Lab, sch.ThemeZip, sch.BaseURL, sch.Root)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scheduler %s enqueue: %v\n", sch.ID, err)
			continue
		}
		fmt.Fprintf(os.Stderr, "scheduler fired %s preset=%s run=%s\n", sch.ID, sch.Preset, runID)
		last := now
		sch.LastRunAt = &last
		sch.NextRunAt = &next
		_ = s.store.UpdateSchedule(ctx, sch)
	}
}

// enqueuePreset creates a queued run and pushes it to the worker queue.
func (s *server) enqueuePreset(ctx context.Context, preset, lab, themeZip, baseURL, root string) (string, error) {
	m, _, err := presets.Load(s.repoRoot, preset)
	if err != nil {
		return "", err
	}
	presets.ApplyBindings(m, presets.Bindings{
		ThemeZip: themeZip,
		BaseURL:  baseURL,
		Root:     root,
	})
	raw, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	if lab == "" {
		lab = m.Spec.Lab
	}
	run := &runstore.Run{
		Lab:          lab,
		Status:       runstore.StatusQueued,
		ManifestJSON: raw,
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.store.CreateRun(ctx, run); err != nil {
		return "", err
	}
	select {
	case s.queue <- run.ID:
	default:
		return "", fmt.Errorf("worker queue full")
	}
	return run.ID, nil
}

func (s *server) handleListSchedules(w http.ResponseWriter, r *http.Request) {
	list, err := s.store.ListSchedules(r.Context())
	if err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"schedules": list})
}

func (s *server) handleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var req createScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Cron == "" || req.Preset == "" {
		writeErr(w, http.StatusBadRequest, "cron and preset required")
		return
	}
	next, err := scheduler.NextAfter(req.Cron, time.Now().UTC())
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if _, _, err := presets.Load(s.repoRoot, req.Preset); err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	sch := &runstore.Schedule{
		Cron:      req.Cron,
		Preset:    req.Preset,
		Lab:       req.Lab,
		Enabled:   enabled,
		ThemeZip:  req.ThemeZip,
		BaseURL:   req.BaseURL,
		Root:      req.Root,
		NextRunAt: &next,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.store.CreateSchedule(r.Context(), sch); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, sch)
}

func (s *server) handleGetSchedule(w http.ResponseWriter, r *http.Request) {
	sch, err := s.store.GetSchedule(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sch)
}

func (s *server) handleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	sch, err := s.store.GetSchedule(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	var req createScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Cron != "" {
		sch.Cron = req.Cron
	}
	if req.Preset != "" {
		sch.Preset = req.Preset
	}
	if req.Lab != "" {
		sch.Lab = req.Lab
	}
	if req.ThemeZip != "" {
		sch.ThemeZip = req.ThemeZip
	}
	if req.BaseURL != "" {
		sch.BaseURL = req.BaseURL
	}
	if req.Root != "" {
		sch.Root = req.Root
	}
	if req.Enabled != nil {
		sch.Enabled = *req.Enabled
	}
	next, err := scheduler.NextAfter(sch.Cron, time.Now().UTC())
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	sch.NextRunAt = &next
	if err := s.store.UpdateSchedule(r.Context(), sch); err != nil {
		writeErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sch)
}

func (s *server) handleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	if err := s.store.DeleteSchedule(r.Context(), r.PathValue("id")); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s *server) handleFireSchedule(w http.ResponseWriter, r *http.Request) {
	sch, err := s.store.GetSchedule(r.Context(), r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	runID, err := s.enqueuePreset(r.Context(), sch.Preset, sch.Lab, sch.ThemeZip, sch.BaseURL, sch.Root)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	now := time.Now().UTC()
	next, _ := scheduler.NextAfter(sch.Cron, now)
	sch.LastRunAt = &now
	sch.NextRunAt = &next
	_ = s.store.UpdateSchedule(r.Context(), sch)
	writeJSON(w, http.StatusAccepted, map[string]any{"runId": runID, "schedule": sch})
}
