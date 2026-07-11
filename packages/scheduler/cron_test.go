package scheduler_test

import (
	"testing"
	"time"

	"github.com/fastygo/lab/packages/scheduler"
)

func TestParseAndDue(t *testing.T) {
	t.Parallel()
	if _, err := scheduler.Parse("not a cron"); err == nil {
		t.Fatal("expected error")
	}
	// every minute
	now := time.Date(2026, 7, 11, 12, 0, 30, 0, time.UTC)
	last := time.Date(2026, 7, 11, 11, 59, 0, 0, time.UTC)
	due, next, err := scheduler.IsDue("* * * * *", &last, nil, now)
	if err != nil {
		t.Fatal(err)
	}
	if !due {
		t.Fatal("expected due")
	}
	if next.Before(now) {
		t.Fatalf("next=%v", next)
	}
	future := now.Add(2 * time.Minute)
	nextAt := now.Add(1 * time.Minute)
	due, _, err = scheduler.IsDue("0 * * * *", nil, &nextAt, future)
	if err != nil {
		t.Fatal(err)
	}
	// nextAt was 12:01, future is 12:02 → due
	if !due {
		t.Fatal("expected due after nextAt")
	}
}
