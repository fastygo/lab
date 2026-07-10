package docker

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

func TestRunnerParsesFindings(t *testing.T) {
	t.Parallel()
	r := New("lighthouse", "lab/lighthouse:local", WithExec(func(ctx context.Context, name string, args []string, env []string) ([]byte, []byte, error) {
		return []byte(`{"findings":[{"code":"quality.lighthouse.performance","severity":"info","message":"score 95"}]}`), nil, nil
	}))
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "Q1",
		Check: domain.Check{ID: "lh", Runner: "lighthouse"},
		Target: domain.Target{BaseURL: "http://example.test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Code != "quality.lighthouse.performance" {
		t.Fatalf("%+v", got)
	}
	if got[0].Gate != "Q1" || got[0].Check != "lh" {
		t.Fatalf("enrich failed: %+v", got[0])
	}
}

func TestRunnerDockerMissing(t *testing.T) {
	t.Parallel()
	r := New("axe", "lab/axe:local", WithExec(func(ctx context.Context, name string, args []string, env []string) ([]byte, []byte, error) {
		return nil, nil, exec.ErrNotFound
	}))
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "Q4",
		Check:  domain.Check{ID: "axe", Runner: "axe"},
		Target: domain.Target{BaseURL: "http://example.test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Code != "runner.docker.unavailable" {
		t.Fatalf("%+v", got)
	}
	if got[0].Severity != domain.SeverityHigh {
		t.Fatalf("severity=%s", got[0].Severity)
	}
}

func TestRunnerThemeZipMountArgs(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	zip := dir + "/theme.zip"
	if err := os.WriteFile(zip, []byte("PK"), 0o644); err != nil {
		t.Fatal(err)
	}
	var sawArgs []string
	r := New("theme-check", "lab/theme-check:local", WithExec(func(ctx context.Context, name string, args []string, env []string) ([]byte, []byte, error) {
		sawArgs = append([]string{}, args...)
		return []byte(`{"findings":[{"code":"org.themecheck.ok","severity":"info","message":"ok"}]}`), nil, nil
	}))
	_, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:  "C2",
		Check: domain.Check{ID: "tc", Runner: "theme-check", Config: map[string]string{
			"dockerNetwork": "fastygo-lab_lab",
			"wpDataVolume":  "fastygo-lab_wp_org_data",
			"internalUrl":   "http://wordpress",
		}},
		Target: domain.Target{
			BaseURL:  "http://127.0.0.1:8080",
			Metadata: map[string]string{"themeZip": zip},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(sawArgs, " ")
	if !strings.Contains(joined, "/lab/theme.zip:ro") {
		t.Fatalf("missing zip mount: %v", sawArgs)
	}
	if !strings.Contains(joined, "fastygo-lab_wp_org_data:/var/www/html") {
		t.Fatalf("missing wp volume: %v", sawArgs)
	}
	if !strings.Contains(joined, "--network fastygo-lab_lab") {
		t.Fatalf("missing network: %v", sawArgs)
	}
	foundEnv := false
	for i := 0; i < len(sawArgs)-1; i++ {
		if sawArgs[i] == "-e" && strings.HasPrefix(sawArgs[i+1], "LAB_TARGET_URL=http://wordpress") {
			foundEnv = true
		}
	}
	if !foundEnv {
		t.Fatalf("internalUrl not applied: %v", sawArgs)
	}
}

func TestRunnerExecFailedFinding(t *testing.T) {
	t.Parallel()
	r := New("wpscan", "wpscanteam/wpscan", WithExec(func(ctx context.Context, name string, args []string, env []string) ([]byte, []byte, error) {
		return nil, []byte("boom"), errors.New("exit 2")
	}))
	got, err := r.Run(context.Background(), ports.RunnerRequest{
		Gate:   "S2",
		Check:  domain.Check{ID: "wpscan", Runner: "wpscan"},
		Target: domain.Target{BaseURL: "http://wp.test"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Code != "runner.exec.failed" {
		t.Fatalf("%+v", got)
	}
}
