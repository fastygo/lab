package docker

import (
	"context"
	"errors"
	"os/exec"
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
