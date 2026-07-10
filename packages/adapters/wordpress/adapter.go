package wordpress

import (
	"context"
	"fmt"
	"os"

	"github.com/fastygo/lab/packages/domain"
)

// Adapter is a stub WordPress target: config supplies baseUrl and optional themeZip.
type Adapter struct {
	baseURL  string
	themeZip string
}

func New() *Adapter { return &Adapter{} }

func (a *Adapter) ID() string { return "wordpress" }

func (a *Adapter) Capabilities() []string {
	return []string{"http", "html", "a11y", "seo", "wp-theme"}
}

func (a *Adapter) Prepare(_ context.Context, config map[string]string) error {
	a.baseURL = config["baseUrl"]
	if a.baseURL == "" {
		a.baseURL = "http://127.0.0.1:8080"
	}
	a.themeZip = config["themeZip"]
	if a.themeZip != "" {
		if _, err := os.Stat(a.themeZip); err != nil {
			return fmt.Errorf("themeZip: %w", err)
		}
	}
	return nil
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
