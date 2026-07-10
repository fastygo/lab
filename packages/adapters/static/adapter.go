package static

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/fastygo/lab/packages/domain"
)

// Adapter serves a local directory of static HTML for quality labs.
type Adapter struct {
	root    string
	server  *http.Server
	baseURL string
	mu      sync.Mutex
}

func New(root string) *Adapter {
	return &Adapter{root: root}
}

func (a *Adapter) ID() string { return "static" }

func (a *Adapter) Capabilities() []string {
	return []string{"http", "html", "a11y", "seo"}
}

func (a *Adapter) Prepare(_ context.Context, config map[string]string) error {
	if r := config["root"]; r != "" {
		a.root = r
	}
	if a.root == "" {
		return fmt.Errorf("static adapter: root directory required")
	}
	st, err := os.Stat(a.root)
	if err != nil {
		return fmt.Errorf("static adapter root: %w", err)
	}
	if !st.IsDir() {
		return fmt.Errorf("static adapter root is not a directory: %s", a.root)
	}
	return nil
}

func (a *Adapter) Serve(ctx context.Context) (domain.Target, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.server != nil {
		return domain.Target{BaseURL: a.baseURL, Metadata: map[string]string{"adapter": "static"}}, nil
	}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return domain.Target{}, err
	}
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(a.root)))
	a.server = &http.Server{Handler: mux}
	a.baseURL = "http://" + ln.Addr().String()
	go func() { _ = a.server.Serve(ln) }()

	// brief readiness
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/", nil)
		resp, err := http.DefaultClient.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode < 500 {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return domain.Target{
		BaseURL:  a.baseURL,
		Metadata: map[string]string{"adapter": "static", "root": a.root},
	}, nil
}

func (a *Adapter) Matrix(_ context.Context) ([]string, error) {
	base := a.baseURL
	if base == "" {
		base = "http://127.0.0.1"
	}
	return []string{base + "/", base + "/about.html"}, nil
}

func (a *Adapter) Teardown(_ context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err := a.server.Shutdown(ctx)
	a.server = nil
	a.baseURL = ""
	return err
}
