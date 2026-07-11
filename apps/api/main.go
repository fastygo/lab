package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Cycle F skeleton API — healthz only until runstore lands.
// Spec: .project/vps/cycle-f-saas.md
const version = "0.0.1-f0"

func main() {
	addr := envOr("LAB_API_ADDR", ":8090")
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":        true,
			"service":   "lab-api",
			"version":   version,
			"cycle":     "F0",
			"ts":        time.Now().UTC().Format(time.RFC3339),
			"roadmap":   "see .project/vps/cycle-f-saas.md",
		})
	})
	mux.HandleFunc("GET /v1/labs", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"labs": []string{"demo", "quality", "org", "sec", "static-web"},
		})
	})

	fmt.Fprintf(os.Stderr, "lab-api %s listening on %s\n", version, addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
