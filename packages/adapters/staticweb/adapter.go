package staticweb

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fastygo/lab/packages/domain"
)

// Adapter serves a Vite/React/Vue/Svelte (or any static SPA) target.
//
// Modes:
//   - dist (default): serve distDir with optional SPA fallback (no Node required)
//   - preview: run previewCommand (e.g. vite preview) when Node is available
//
// If config baseUrl is set, Serve returns that URL without starting a server
// (useful when preview already runs in Compose).
type Adapter struct {
	repoRoot string

	root           string
	distDir        string
	mode           string
	build          bool
	buildCommand   string
	previewCommand string
	matrixPaths    []string
	spa            bool
	framework      string
	externalURL    string

	server  *http.Server
	cmd     *exec.Cmd
	baseURL string
	mu      sync.Mutex
}

func New(repoRoot string) *Adapter {
	return &Adapter{repoRoot: repoRoot}
}

func (a *Adapter) ID() string { return "static-web" }

func (a *Adapter) Capabilities() []string {
	return []string{"http", "html", "a11y", "seo", "spa"}
}

func (a *Adapter) Prepare(_ context.Context, config map[string]string) error {
	a.externalURL = strings.TrimRight(config["baseUrl"], "/")
	a.mode = firstNonEmpty(config["mode"], "dist")
	a.distDir = firstNonEmpty(config["dist"], "dist")
	a.build = config["build"] == "true" || config["build"] == "1"
	a.buildCommand = firstNonEmpty(config["buildCommand"], "npm run build")
	a.previewCommand = config["previewCommand"]
	a.spa = config["spa"] != "false" && config["spa"] != "0"
	a.framework = firstNonEmpty(config["framework"], "vite")
	a.matrixPaths = splitCSV(firstNonEmpty(config["matrix"], "/,/about,/about.html"))

	root := config["root"]
	if root == "" {
		root = filepath.Join(a.repoRoot, "testdata", "fixtures", "static-web-app")
	}
	if !filepath.IsAbs(root) && a.repoRoot != "" {
		cand := filepath.Join(a.repoRoot, root)
		if st, err := os.Stat(cand); err == nil && st.IsDir() {
			root = cand
		} else if abs, err := filepath.Abs(root); err == nil {
			root = abs
		}
	}
	st, err := os.Stat(root)
	if err != nil {
		return fmt.Errorf("static-web root: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("static-web root is not a directory: %s", root)
	}
	a.root = root

	if a.externalURL != "" {
		return nil
	}

	if a.build {
		if err := a.runShell(a.buildCommand, a.root); err != nil {
			return fmt.Errorf("static-web build: %w", err)
		}
	}

	if a.mode == "dist" || a.mode == "serve-dist" {
		distPath := filepath.Join(a.root, a.distDir)
		if st, err := os.Stat(distPath); err != nil || !st.IsDir() {
			return fmt.Errorf("static-web dist missing: %s (set build=true or provide dist/)", distPath)
		}
	}
	return nil
}

func (a *Adapter) Serve(ctx context.Context) (domain.Target, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	meta := map[string]string{
		"adapter":    "static-web",
		"framework":  a.framework,
		"mode":       a.mode,
		"root":       a.root,
		"dist":       a.distDir,
		"spa":        fmt.Sprintf("%v", a.spa),
	}

	if a.externalURL != "" {
		a.baseURL = a.externalURL
		return domain.Target{BaseURL: a.baseURL, Metadata: meta}, nil
	}
	if a.baseURL != "" && (a.server != nil || a.cmd != nil) {
		return domain.Target{BaseURL: a.baseURL, Metadata: meta}, nil
	}

	switch a.mode {
	case "preview":
		if err := a.servePreview(ctx); err != nil {
			return domain.Target{}, err
		}
	default:
		if err := a.serveDist(ctx); err != nil {
			return domain.Target{}, err
		}
	}
	return domain.Target{BaseURL: a.baseURL, Metadata: meta}, nil
}

func (a *Adapter) serveDist(ctx context.Context) error {
	distPath := filepath.Join(a.root, a.distDir)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	fs := http.FileServer(http.Dir(distPath))
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if a.spa {
			serveSPA(w, r, distPath, fs)
			return
		}
		fs.ServeHTTP(w, r)
	})
	a.server = &http.Server{Handler: mux}
	a.baseURL = "http://" + ln.Addr().String()
	go func() { _ = a.server.Serve(ln) }()
	return waitReady(ctx, a.baseURL+"/")
}

func (a *Adapter) servePreview(ctx context.Context) error {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()

	cmdStr := a.previewCommand
	if cmdStr == "" {
		cmdStr = fmt.Sprintf("npx --yes vite preview --host 127.0.0.1 --port %d --strictPort", port)
	} else {
		cmdStr = strings.ReplaceAll(cmdStr, "{port}", fmt.Sprintf("%d", port))
	}

	cmd, err := shellCmd(cmdStr, a.root)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("static-web preview start: %w", err)
	}
	a.cmd = cmd
	a.baseURL = fmt.Sprintf("http://127.0.0.1:%d", port)
	if err := waitReady(ctx, a.baseURL+"/"); err != nil {
		_ = a.stopPreview()
		return err
	}
	return nil
}

func (a *Adapter) Matrix(_ context.Context) ([]string, error) {
	base := a.baseURL
	if base == "" {
		base = "http://127.0.0.1"
	}
	out := make([]string, 0, len(a.matrixPaths))
	for _, p := range a.matrixPaths {
		if !strings.HasPrefix(p, "/") {
			p = "/" + p
		}
		out = append(out, base+p)
	}
	return out, nil
}

func (a *Adapter) Teardown(_ context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	var err error
	if a.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = a.server.Shutdown(ctx)
		cancel()
		a.server = nil
	}
	if a.cmd != nil {
		if e := a.stopPreview(); e != nil && err == nil {
			err = e
		}
	}
	a.baseURL = ""
	return err
}

func (a *Adapter) stopPreview() error {
	if a.cmd == nil || a.cmd.Process == nil {
		a.cmd = nil
		return nil
	}
	_ = a.cmd.Process.Kill()
	_, _ = a.cmd.Process.Wait()
	a.cmd = nil
	return nil
}

func serveSPA(w http.ResponseWriter, r *http.Request, dist string, fs http.Handler) {
	path := r.URL.Path
	if path == "/" {
		fs.ServeHTTP(w, r)
		return
	}
	clean := filepath.Clean("/" + path)
	full := filepath.Join(dist, filepath.FromSlash(clean))
	if st, err := os.Stat(full); err == nil && !st.IsDir() {
		fs.ServeHTTP(w, r)
		return
	}
	// Directory with index.html
	if st, err := os.Stat(full); err == nil && st.IsDir() {
		idx := filepath.Join(full, "index.html")
		if _, err := os.Stat(idx); err == nil {
			fs.ServeHTTP(w, r)
			return
		}
	}
	// Extension present → real 404 from file server
	if ext := filepath.Ext(clean); ext != "" && ext != ".html" {
		fs.ServeHTTP(w, r)
		return
	}
	http.ServeFile(w, r, filepath.Join(dist, "index.html"))
}

func waitReady(ctx context.Context, url string) error {
	deadline := time.Now().Add(8 * time.Second)
	var last error
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 500 {
				return nil
			}
			last = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			last = err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(40 * time.Millisecond):
		}
	}
	if last == nil {
		last = fmt.Errorf("timeout")
	}
	return fmt.Errorf("static-web not ready: %w", last)
}

func (a *Adapter) runShell(command, dir string) error {
	cmd, err := shellCmd(command, dir)
	if err != nil {
		return err
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func shellCmd(command, dir string) (*exec.Cmd, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, fmt.Errorf("empty command")
	}
	var cmd *exec.Cmd
	if _, err := exec.LookPath("bash"); err == nil {
		cmd = exec.Command("bash", "-lc", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = dir
	return cmd, nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
