package scheduler

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
)

var parser = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)

// Parse validates a 5-field cron expression.
func Parse(expr string) (cron.Schedule, error) {
	s, err := parser.Parse(expr)
	if err != nil {
		return nil, fmt.Errorf("cron %q: %w", expr, err)
	}
	return s, nil
}

// NextAfter returns the next fire time after t.
func NextAfter(expr string, t time.Time) (time.Time, error) {
	s, err := Parse(expr)
	if err != nil {
		return time.Time{}, err
	}
	return s.Next(t), nil
}

// IsDue reports whether the schedule should fire at now.
// Uses NextRunAt when set; otherwise computes from LastRunAt (or epoch).
func IsDue(cronExpr string, lastRun, nextRun *time.Time, now time.Time) (bool, time.Time, error) {
	s, err := Parse(cronExpr)
	if err != nil {
		return false, time.Time{}, err
	}
	now = now.UTC()
	var dueAt time.Time
	if nextRun != nil && !nextRun.IsZero() {
		dueAt = nextRun.UTC()
	} else {
		base := time.Unix(0, 0).UTC()
		if lastRun != nil && !lastRun.IsZero() {
			base = lastRun.UTC()
		}
		dueAt = s.Next(base)
	}
	if now.Before(dueAt) {
		return false, s.Next(now), nil
	}
	return true, s.Next(now), nil
}
