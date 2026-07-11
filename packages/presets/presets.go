// Package presets maps SaaS preset ids to Manifest files under the repo.
package presets

import (
	"fmt"
	"path/filepath"

	"github.com/fastygo/lab/packages/domain"
)

// Known preset ids for Cycle F jobs.
var Known = map[string]string{
	"demo":             "testdata/manifests/demo.lab.yaml",
	"quality":          "testdata/manifests/quality.lab.yaml",
	"quality-wp":       "testdata/manifests/quality-wp.lab.yaml",
	"org":              "testdata/manifests/org.lab.yaml",
	"sec":              "testdata/manifests/sec.lab.yaml",
	"static-web":       "testdata/manifests/quality-staticweb.lab.yaml",
	"quality-staticweb": "testdata/manifests/quality-staticweb.lab.yaml",
}

// Load resolves a preset name (or explicit relative path) to a Manifest.
func Load(repoRoot, presetOrPath string) (*domain.Manifest, string, error) {
	rel, ok := Known[presetOrPath]
	if !ok {
		// allow direct path under repo for power users
		rel = presetOrPath
	}
	path := rel
	if !filepath.IsAbs(path) {
		path = filepath.Join(repoRoot, rel)
	}
	m, err := domain.LoadManifest(path)
	if err != nil {
		return nil, "", fmt.Errorf("preset %q: %w", presetOrPath, err)
	}
	return m, path, nil
}

// List returns preset ids.
func List() []string {
	out := make([]string, 0, len(Known))
	for k := range Known {
		out = append(out, k)
	}
	return out
}
