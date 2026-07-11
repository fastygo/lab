package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/fastygo/lab/apps/web/views"
	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/reportfmt"
)

// Cycle F dashboard — SSR templ over lab-api JSON.
const version = "0.1.0-f4"

type apiClient struct {
	base   string
	client *http.Client
}

func main() {
	addr := envOr("LAB_WEB_ADDR", ":8091")
	apiBase := envOr("LAB_API_URL", "http://127.0.0.1:8090")
	api := &apiClient{
		base:   stringsTrimRightSlash(apiBase),
		client: &http.Client{Timeout: 30 * time.Second},
	}
	staticDir := filepath.Join(findWebRoot(), "static")

	mux := http.NewServeMux()
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		lab := r.URL.Query().Get("lab")
		runs, err := api.listRuns(r.Context(), lab, 50)
		props := views.RunsPageProps{APIBase: api.base, Lab: lab, Runs: runs}
		if err != nil {
			props.Err = err.Error()
		}
		if err := views.RunsPage(props).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("GET /runs/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		run, err := api.getRun(r.Context(), id)
		props := views.RunDetailProps{APIBase: api.base, Run: run}
		if err != nil {
			props.Err = err.Error()
		} else {
			props.Live = views.IsLiveStatus(run.Status)
			if report, rerr := api.getReport(r.Context(), id); rerr == nil {
				props.Report = report
				if run.Summary == nil {
					props.Run.Summary = &report.Summary
				}
				props.Baskets = views.CountBaskets(report)
			}
			if events, eerr := api.getEvents(r.Context(), id); eerr == nil {
				props.Events = events
			}
		}
		if err := views.RunDetailPage(props).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("GET /runs/{id}/live", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		q := r.URL.RawQuery
		path := "/v1/runs/" + url.PathEscape(id) + "/events/stream"
		if q != "" {
			path += "?" + q
		}
		if err := api.proxySSE(r.Context(), w, path); err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
		}
	})
	mux.HandleFunc("GET /compare", func(w http.ResponseWriter, r *http.Request) {
		baseID := r.URL.Query().Get("base")
		headID := r.URL.Query().Get("head")
		runs, err := api.listRuns(r.Context(), "", 100)
		props := views.ComparePageProps{APIBase: api.base, BaseID: baseID, HeadID: headID, Runs: runs}
		if err != nil {
			props.Err = err.Error()
		} else if baseID != "" && headID != "" {
			diff, derr := api.compare(r.Context(), baseID, headID)
			if derr != nil {
				props.Err = derr.Error()
			} else {
				props.Diff = diff
			}
		}
		if err := views.ComparePage(props).Render(r.Context(), w); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok": true, "service": "lab-web", "version": version, "api": api.base,
		})
	})

	fmt.Fprintf(os.Stderr, "lab-web %s listening on %s (api=%s)\n", version, addr, api.base)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func (a *apiClient) listRuns(ctx context.Context, lab string, limit int) ([]views.RunRow, error) {
	q := url.Values{}
	if lab != "" {
		q.Set("lab", lab)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := "/v1/runs"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	var wrap struct {
		Runs []views.RunRow `json:"runs"`
	}
	if err := a.getJSON(ctx, path, &wrap); err != nil {
		return nil, err
	}
	return wrap.Runs, nil
}

func (a *apiClient) getRun(ctx context.Context, id string) (views.RunRow, error) {
	var run views.RunRow
	err := a.getJSON(ctx, "/v1/runs/"+url.PathEscape(id), &run)
	return run, err
}

func (a *apiClient) getReport(ctx context.Context, id string) (*domain.Report, error) {
	var report domain.Report
	if err := a.getJSON(ctx, "/v1/runs/"+url.PathEscape(id)+"/report", &report); err != nil {
		return nil, err
	}
	return &report, nil
}

func (a *apiClient) getEvents(ctx context.Context, id string) ([]domain.RunEvent, error) {
	var wrap struct {
		Events []domain.RunEvent `json:"events"`
	}
	if err := a.getJSON(ctx, "/v1/runs/"+url.PathEscape(id)+"/events", &wrap); err != nil {
		return nil, err
	}
	return wrap.Events, nil
}

func (a *apiClient) compare(ctx context.Context, baseID, headID string) (*reportfmt.Diff, error) {
	q := url.Values{"base": {baseID}, "head": {headID}}
	var diff reportfmt.Diff
	if err := a.getJSON(ctx, "/v1/runs/compare?"+q.Encode(), &diff); err != nil {
		return nil, err
	}
	return &diff, nil
}

func (a *apiClient) getJSON(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.base+path, nil)
	if err != nil {
		return err
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("api %s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(dest)
}

// proxySSE forwards an API SSE stream to the browser (same-origin EventSource).
func (a *apiClient) proxySSE(ctx context.Context, w http.ResponseWriter, path string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.base+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")
	client := &http.Client{Timeout: 0} // long-lived stream
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("api stream: %s", resp.Status)
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()
	buf := make([]byte, 4096)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			if _, werr := w.Write(buf[:n]); werr != nil {
				return nil
			}
			flusher.Flush()
		}
		if err != nil {
			return nil
		}
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func stringsTrimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func findWebRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "apps/web"
	}
	candidates := []string{
		filepath.Join(wd, "apps", "web"),
		wd,
	}
	for _, c := range candidates {
		if st, err := os.Stat(filepath.Join(c, "static", "css", "app.css")); err == nil && !st.IsDir() {
			return c
		}
	}
	return filepath.Join(wd, "apps", "web")
}
