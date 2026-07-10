package wordpress

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fastygo/lab/packages/domain"
)

// Adapter is a stub WordPress target: config supplies baseUrl and optional themeZip.
type Adapter struct {
	repoRoot string
	baseURL  string
	themeZip string
}

func New(repoRoot string) *Adapter {
	return &Adapter{repoRoot: repoRoot}
}

func (a *Adapter) ID() string { return "wordpress" }

func (a *Adapter) Capabilities() []string {
	return []string{"http", "html", "a11y", "seo", "wp-theme"}
}

func (a *Adapter) Prepare(_ context.Context, config map[string]string) error {
	a.baseURL = config["baseUrl"]
	if a.baseURL == "" {
		a.baseURL = "http://127.0.0.1:8080"
	}
	zip := config["themeZip"]
	if zip != "" {
		resolved, err := resolveThemeZip(zip, a.repoRoot)
		if err != nil {
			return fmt.Errorf("themeZip: %w", err)
		}
		a.themeZip = resolved
	}
	return nil
}

func resolveThemeZip(zip, repoRoot string) (string, error) {
	candidates := []string{zip}
	if !filepath.IsAbs(zip) {
		if wd, err := os.Getwd(); err == nil {
			candidates = append(candidates, filepath.Join(wd, zip))
		}
		if repoRoot != "" {
			candidates = append(candidates, filepath.Join(repoRoot, zip))
		}
	}
	var last error
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			last = err
			continue
		}
		if st, err := os.Stat(abs); err == nil && !st.IsDir() {
			return abs, nil
		} else if err != nil {
			last = err
		}
	}
	if last == nil {
		last = os.ErrNotExist
	}
	return "", last
}

func (a *Adapter) Serve(_ context.Context) (domain.Target, error) {
	meta := map[string]string{"adapter": "wordpress"}
	if a.themeZip != "" {
		meta["themeZip"] = a.themeZip
	}
	return domain.Target{BaseURL: a.baseURL, Metadata: meta}, nil
}

func (a *Adapter) Matrix(_ context.Context) ([]string, error) {
	b := a.baseURL
	return []string{
		b + "/",
		b + "/?p=1",
		b + "/sample-page/",
		b + "/category/uncategorized/",
		b + "/tag/test/",
		b + "/author/admin/",
		b + "/?s=hello",
		b + "/this-page-does-not-exist-lab-404/",
	}, nil
}

func (a *Adapter) Teardown(_ context.Context) error { return nil }
