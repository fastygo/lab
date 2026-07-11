// Package registry wires default adapters and runners for the CLI.
package registry

import (
	"os"
	"path/filepath"

	"github.com/fastygo/lab/packages/adapters/noop"
	"github.com/fastygo/lab/packages/adapters/static"
	"github.com/fastygo/lab/packages/adapters/wordpress"
	"github.com/fastygo/lab/packages/orchestrator"
	"github.com/fastygo/lab/packages/orchestrator/docker"
	"github.com/fastygo/lab/packages/orchestrator/headers"
	"github.com/fastygo/lab/packages/orchestrator/httpmatrix"
	"github.com/fastygo/lab/packages/orchestrator/stub"
	"github.com/fastygo/lab/packages/orchestrator/ziplint"
)

// DefaultAdapters returns built-in target adapters.
func DefaultAdapters(repoRoot string) []orchestrator.TargetAdapter {
	fixture := filepath.Join(repoRoot, "testdata", "fixtures", "quality-site")
	return []orchestrator.TargetAdapter{
		noop.New(),
		static.New(fixture),
		wordpress.New(repoRoot),
	}
}

// DefaultRunners returns built-in runners (in-process + docker-backed).
func DefaultRunners() []orchestrator.Runner {
	lhImage := envOr("LAB_LIGHTHOUSE_IMAGE", "lab/lighthouse:local")
	axeImage := envOr("LAB_AXE_IMAGE", "lab/axe:local")
	tcImage := envOr("LAB_THEMECHECK_IMAGE", "lab/theme-check:local")
	vnuImage := envOr("LAB_VNU_IMAGE", "lab/vnu:local")
	wpscanImage := envOr("LAB_WPSCAN_IMAGE", "wpscanteam/wpscan:latest")
	noticeImage := envOr("LAB_NOTICE_HUNTER_IMAGE", "lab/notice-hunter:local")

	return []orchestrator.Runner{
		stub.New(),
		ziplint.New(),
		headers.New(),
		httpmatrix.New(),
		docker.New("lighthouse", lhImage),
		docker.New("axe", axeImage),
		docker.New("theme-check", tcImage),
		docker.New("vnu", vnuImage),
		docker.New("wpscan", wpscanImage),
		docker.New("notice-hunter", noticeImage),
	}
}

// KnownLabs lists product lab ids.
func KnownLabs() []string {
	return []string{"demo", "quality", "org", "sec"}
}

// FindRepoRoot walks up from cwd looking for go.mod.
func FindRepoRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}
		dir = parent
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
