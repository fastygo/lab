package docker

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/orchestrator/ports"
)

// ExecFunc runs an external command (injectable for tests).
type ExecFunc func(ctx context.Context, name string, args []string, env []string) (stdout, stderr []byte, err error)

// Runner invokes a container image and parses findings JSON from stdout.
type Runner struct {
	id        string
	image     string
	docker    string
	exec      ExecFunc
	extraArgs []string
}

// Option configures a Runner.
type Option func(*Runner)

func WithExec(fn ExecFunc) Option {
	return func(r *Runner) { r.exec = fn }
}

func WithDockerBin(bin string) Option {
	return func(r *Runner) { r.docker = bin }
}

func WithExtraArgs(args ...string) Option {
	return func(r *Runner) { r.extraArgs = append(r.extraArgs, args...) }
}

// New creates a docker-backed runner. id is the runner id used in manifests (e.g. "lighthouse").
func New(id, image string, opts ...Option) *Runner {
	r := &Runner{
		id:     id,
		image:  image,
		docker: "docker",
		exec:   defaultExec,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Runner) ID() string { return r.id }

func (r *Runner) Run(ctx context.Context, req ports.RunnerRequest) ([]domain.Finding, error) {
	image := req.Check.Config["image"]
	if image == "" {
		image = r.image
	}
	if image == "" {
		return []domain.Finding{unavailable(req, "image not configured")}, nil
	}

	targetURL := req.Target.BaseURL
	if u := firstNonEmpty(req.Check.Config["internalUrl"], os.Getenv("LAB_INTERNAL_URL")); u != "" {
		targetURL = u
	}

	cfgJSON, _ := json.Marshal(req.Check.Config)
	env := []string{
		"LAB_TARGET_URL=" + targetURL,
		"LAB_GATE_ID=" + req.Gate,
		"LAB_CHECK_ID=" + req.Check.ID,
		"LAB_CONFIG_JSON=" + string(cfgJSON),
	}
	if tok := os.Getenv("WPSCAN_API_TOKEN"); tok != "" {
		env = append(env, "WPSCAN_API_TOKEN="+tok)
	}

	args := []string{"run", "--rm"}
	args = append(args, r.extraArgs...)

	// Theme zip: mount host file into the container at a stable path.
	zipHost := firstNonEmpty(req.Check.Config["themeZip"], req.Target.Metadata["themeZip"])
	if zipHost != "" {
		abs, err := filepath.Abs(zipHost)
		if err == nil {
			zipHost = abs
		}
		args = append(args, "-v", zipHost+":/lab/theme.zip:ro")
		env = append(env, "LAB_THEME_ZIP=/lab/theme.zip")
	}

	// Optional WP data volume (compose org profile) so theme-check shares files with wordpress.
	if vol := firstNonEmpty(req.Check.Config["wpDataVolume"], os.Getenv("LAB_WP_DATA_VOLUME")); vol != "" {
		args = append(args, "-v", vol+":/var/www/html")
	}

	network := firstNonEmpty(req.Check.Config["dockerNetwork"], os.Getenv("LAB_DOCKER_NETWORK"), "host")
	args = append(args, "--network", network)

	for _, e := range env {
		args = append(args, "-e", e)
	}
	// Pass-through extra env from check config: env.FOO=bar → -e FOO=bar
	for k, v := range req.Check.Config {
		if strings.HasPrefix(k, "env.") {
			args = append(args, "-e", strings.TrimPrefix(k, "env.")+"="+v)
		}
	}
	args = append(args, image)

	stdout, stderr, err := r.exec(ctx, r.docker, args, env)
	if err != nil {
		if isDockerMissing(err, stderr) {
			return []domain.Finding{unavailable(req, err.Error())}, nil
		}
		// Tool ran but failed: still try parse stdout; else emit runner error finding.
		if findings, perr := parseFindings(stdout, req); perr == nil && len(findings) > 0 {
			return findings, nil
		}
		return []domain.Finding{{
			Code:     "runner.exec.failed",
			Gate:     req.Gate,
			Check:    req.Check.ID,
			Severity: domain.SeverityHigh,
			Message:  fmt.Sprintf("runner %s failed: %v: %s", r.id, err, strings.TrimSpace(string(stderr))),
			Target:   req.Target.BaseURL,
		}}, nil
	}
	return parseFindings(stdout, req)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func unavailable(req ports.RunnerRequest, detail string) domain.Finding {
	return domain.Finding{
		Code:     "runner.docker.unavailable",
		Gate:     req.Gate,
		Check:    req.Check.ID,
		Severity: domain.SeverityHigh,
		Message:  "Docker runner unavailable: " + detail,
		Target:   req.Target.BaseURL,
		Evidence: map[string]string{"runner": req.Check.Runner},
	}
}

func isDockerMissing(err error, stderr []byte) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error() + " " + string(stderr))
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}
	return strings.Contains(msg, "executable file not found") ||
		strings.Contains(msg, "cannot find the file") ||
		strings.Contains(msg, "docker desktop") && strings.Contains(msg, "not running") ||
		strings.Contains(msg, "error during connect")
}

func parseFindings(stdout []byte, req ports.RunnerRequest) ([]domain.Finding, error) {
	stdout = bytes.TrimSpace(stdout)
	if len(stdout) == 0 {
		return nil, fmt.Errorf("empty runner stdout")
	}
	var wrap struct {
		Findings []domain.Finding `json:"findings"`
	}
	if err := json.Unmarshal(stdout, &wrap); err == nil && wrap.Findings != nil {
		return enrich(wrap.Findings, req), nil
	}
	var list []domain.Finding
	if err := json.Unmarshal(stdout, &list); err != nil {
		return nil, fmt.Errorf("parse findings: %w", err)
	}
	return enrich(list, req), nil
}

func enrich(findings []domain.Finding, req ports.RunnerRequest) []domain.Finding {
	for i := range findings {
		if findings[i].Gate == "" {
			findings[i].Gate = req.Gate
		}
		if findings[i].Check == "" {
			findings[i].Check = req.Check.ID
		}
		if findings[i].Target == "" {
			findings[i].Target = req.Target.BaseURL
		}
	}
	return findings
}

func defaultExec(ctx context.Context, name string, args []string, env []string) (stdout, stderr []byte, err error) {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Env = append(os.Environ(), env...)
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err = cmd.Run()
	return outBuf.Bytes(), errBuf.Bytes(), err
}
