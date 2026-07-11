package main

import (
	"net/http"

	"github.com/fastygo/lab/packages/reportfmt"
)

func (s *server) handleGetReportMarkdown(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, err := s.store.GetRun(r.Context(), id)
	if err != nil {
		writeErr(w, http.StatusNotFound, err.Error())
		return
	}
	md := reportfmt.Markdown(run, run.Report)
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="lab-`+id[:min(8, len(id))]+`.md"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(md))
}

func (s *server) handleCompareRuns(w http.ResponseWriter, r *http.Request) {
	baseID := r.URL.Query().Get("base")
	headID := r.URL.Query().Get("head")
	if baseID == "" || headID == "" {
		writeErr(w, http.StatusBadRequest, "query base and head run ids required")
		return
	}
	base, err := s.store.GetRun(r.Context(), baseID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "base: "+err.Error())
		return
	}
	head, err := s.store.GetRun(r.Context(), headID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "head: "+err.Error())
		return
	}
	if base.Report == nil || head.Report == nil {
		writeErr(w, http.StatusConflict, "both runs need a finished report")
		return
	}
	diff := reportfmt.CompareReports(baseID, headID, base.Report, head.Report)
	if r.URL.Query().Get("format") == "md" || r.Header.Get("Accept") == "text/markdown" {
		md := reportfmt.DiffMarkdown(diff)
		w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(md))
		return
	}
	writeJSON(w, http.StatusOK, diff)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
