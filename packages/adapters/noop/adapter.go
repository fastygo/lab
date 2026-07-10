package noop

import (
	"context"

	"github.com/fastygo/lab/packages/domain"
)

// Adapter is a fixture target adapter for demo/TDD.
type Adapter struct {
	baseURL string
	matrix  []string
}

func New() *Adapter {
	return &Adapter{
		baseURL: "http://127.0.0.1:9",
		matrix:  []string{"http://127.0.0.1:9/", "http://127.0.0.1:9/health"},
	}
}

func (a *Adapter) ID() string { return "noop" }

func (a *Adapter) Capabilities() []string {
	return []string{"http", "html"}
}

func (a *Adapter) Prepare(_ context.Context, config map[string]string) error {
	if u := config["baseUrl"]; u != "" {
		a.baseURL = u
		a.matrix = []string{u + "/", u + "/health"}
	}
	return nil
}

func (a *Adapter) Serve(_ context.Context) (domain.Target, error) {
	return domain.Target{
		BaseURL: a.baseURL,
		Metadata: map[string]string{
			"adapter": "noop",
		},
	}, nil
}

func (a *Adapter) Matrix(_ context.Context) ([]string, error) {
	return append([]string(nil), a.matrix...), nil
}

func (a *Adapter) Teardown(_ context.Context) error { return nil }
