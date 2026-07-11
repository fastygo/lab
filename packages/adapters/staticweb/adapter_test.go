package staticweb_test

import (
	"context"
	"io"
	"net/http"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/fastygo/lab/packages/adapters/staticweb"
)

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "testdata", "fixtures", "static-web-app"))
}

func TestStaticWebDistSPA(t *testing.T) {
	t.Parallel()
	a := staticweb.New("")
	if err := a.Prepare(context.Background(), map[string]string{
		"root":   fixtureRoot(t),
		"mode":   "dist",
		"matrix": "/,/about,/missing-route",
	}); err != nil {
		t.Fatal(err)
	}
	target, err := a.Serve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = a.Teardown(context.Background()) })

	resp, err := http.Get(target.BaseURL + "/")
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "Static Web") {
		t.Fatalf("home status=%d body=%q", resp.StatusCode, body)
	}

	// Real about/index.html
	resp, err = http.Get(target.BaseURL + "/about")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("about status=%d", resp.StatusCode)
	}

	// SPA fallback for client route
	resp, err = http.Get(target.BaseURL + "/missing-route")
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 || !strings.Contains(string(body), "Home") {
		t.Fatalf("spa fallback status=%d body=%q", resp.StatusCode, body)
	}

	urls, err := a.Matrix(context.Background())
	if err != nil || len(urls) != 3 {
		t.Fatalf("matrix=%v err=%v", urls, err)
	}
}
