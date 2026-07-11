package notify_test

import (
	"strings"
	"testing"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/notify"
	"github.com/fastygo/lab/packages/runstore"
)

func TestShouldSend(t *testing.T) {
	t.Parallel()
	cfg := notify.Config{Filter: notify.FilterWarnFail}
	if !cfg.ShouldSend(runstore.StatusFail) || !cfg.ShouldSend(runstore.StatusWarn) {
		t.Fatal("warn+fail")
	}
	if cfg.ShouldSend(runstore.StatusPass) {
		t.Fatal("pass should skip")
	}
	cfg.Filter = notify.FilterAlways
	if !cfg.ShouldSend(runstore.StatusPass) {
		t.Fatal("always")
	}
}

func TestMessage(t *testing.T) {
	t.Parallel()
	run := &runstore.Run{
		ID:     "abc",
		Lab:    "sec",
		Status: runstore.StatusFail,
		Report: &domain.Report{Summary: domain.Summary{High: 2, Medium: 3, Total: 10}},
	}
	msg := notify.Message(run, "http://lab.example")
	if !strings.Contains(msg, "[sec]") || !strings.Contains(msg, "/runs/abc") {
		t.Fatalf("msg=%q", msg)
	}
}
