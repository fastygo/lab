package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/fastygo/lab/packages/runstore"
)

// handleEventsStream streams run events as Server-Sent Events (F3.4).
// Polls the store so it works for memory and Postgres without a push hub.
func (s *server) handleEventsStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := s.store.GetRun(r.Context(), id); err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sent := 0
	if v := r.URL.Query().Get("after"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			sent = n
		}
	}
	ticker := time.NewTicker(400 * time.Millisecond)
	defer ticker.Stop()
	heartbeat := time.NewTicker(15 * time.Second)
	defer heartbeat.Stop()

	writeSSE := func(event string, v any) error {
		raw, err := json.Marshal(v)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, raw); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprintf(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case <-ticker.C:
			events, err := s.store.ListEvents(r.Context(), id)
			if err != nil {
				_ = writeSSE("error", map[string]string{"error": err.Error()})
				return
			}
			for sent < len(events) {
				if err := writeSSE("event", events[sent]); err != nil {
					return
				}
				sent++
			}
			run, err := s.store.GetRun(r.Context(), id)
			if err != nil {
				_ = writeSSE("error", map[string]string{"error": err.Error()})
				return
			}
			_ = writeSSE("status", map[string]any{
				"id":     run.ID,
				"status": run.Status,
				"error":  run.Error,
			})
			if isTerminal(run.Status) {
				_ = writeSSE("done", map[string]any{
					"id":     run.ID,
					"status": run.Status,
				})
				return
			}
		}
	}
}

func isTerminal(st runstore.RunStatus) bool {
	switch st {
	case runstore.StatusPass, runstore.StatusWarn, runstore.StatusFail, runstore.StatusError:
		return true
	default:
		return false
	}
}
