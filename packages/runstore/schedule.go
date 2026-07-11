package runstore

import (
	"context"
	"time"
)

// Schedule is a cron job that enqueues a lab preset (Cycle F4).
type Schedule struct {
	ID        string     `json:"id"`
	Cron      string     `json:"cron"` // standard 5-field (min hour dom month dow)
	Preset    string     `json:"preset"`
	Lab       string     `json:"lab,omitempty"`
	Enabled   bool       `json:"enabled"`
	ThemeZip  string     `json:"themeZip,omitempty"`
	BaseURL   string     `json:"baseUrl,omitempty"`
	Root      string     `json:"root,omitempty"`
	LastRunAt *time.Time `json:"lastRunAt,omitempty"`
	NextRunAt *time.Time `json:"nextRunAt,omitempty"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// ScheduleStore persists cron schedules.
type ScheduleStore interface {
	CreateSchedule(ctx context.Context, s *Schedule) error
	GetSchedule(ctx context.Context, id string) (*Schedule, error)
	ListSchedules(ctx context.Context) ([]*Schedule, error)
	UpdateSchedule(ctx context.Context, s *Schedule) error
	DeleteSchedule(ctx context.Context, id string) error
}
