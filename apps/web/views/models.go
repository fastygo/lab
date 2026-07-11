package views

import (
	"fmt"
	"strconv"
	"time"

	"github.com/fastygo/lab/packages/domain"
)

// RunRow is one row from GET /v1/runs (and detail extras).
type RunRow struct {
	ID           string          `json:"id"`
	Lab          string          `json:"lab"`
	Status       string          `json:"status"`
	CreatedAt    time.Time       `json:"createdAt"`
	StartedAt    *time.Time      `json:"startedAt,omitempty"`
	FinishedAt   *time.Time      `json:"finishedAt,omitempty"`
	ReportStatus string          `json:"reportStatus,omitempty"`
	Summary      *domain.Summary `json:"summary,omitempty"`
	Error        string          `json:"error,omitempty"`
	EventCount   int             `json:"eventCount,omitempty"`
	LastEvent    string          `json:"lastEvent,omitempty"`
}

type RunsPageProps struct {
	APIBase string
	Lab     string
	Runs    []RunRow
	Err     string
}

type RunDetailProps struct {
	APIBase  string
	Run      RunRow
	Report   *domain.Report
	Events   []domain.RunEvent
	Err      string
	Baskets  []BasketCount
}

type BasketCount struct {
	Basket string
	Count  int
}

func StatusVariant(status string) string {
	switch status {
	case "pass":
		return "default"
	case "warn":
		return "secondary"
	case "fail", "error":
		return "destructive"
	case "running", "queued":
		return "outline"
	default:
		return "outline"
	}
}

func SeverityVariant(sev domain.Severity) string {
	switch sev {
	case domain.SeverityCritical, domain.SeverityHigh:
		return "destructive"
	case domain.SeverityMedium:
		return "secondary"
	default:
		return "outline"
	}
}

func FormatTime(t time.Time) string {
	if t.IsZero() {
		return "—"
	}
	return t.UTC().Format("2006-01-02 15:04:05") + "Z"
}

func FormatPtrTime(t *time.Time) string {
	if t == nil {
		return "—"
	}
	return FormatTime(*t)
}

func Duration(start, end *time.Time) string {
	if start == nil {
		return "—"
	}
	fin := time.Now().UTC()
	if end != nil {
		fin = *end
	}
	d := fin.Sub(*start).Round(time.Second)
	if d < 0 {
		return "—"
	}
	return d.String()
}

func ShortID(id string) string {
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}

func FormatInt(n int) string {
	return strconv.Itoa(n)
}

func CountBaskets(report *domain.Report) []BasketCount {
	if report == nil {
		return nil
	}
	m := map[string]int{}
	order := []string{}
	for _, d := range report.Decisions {
		b := string(d.Basket)
		if _, ok := m[b]; !ok {
			order = append(order, b)
		}
		m[b]++
	}
	out := make([]BasketCount, 0, len(order))
	for _, b := range order {
		out = append(out, BasketCount{Basket: b, Count: m[b]})
	}
	return out
}

func EventLabel(ev domain.RunEvent) string {
	switch string(ev.Type) {
	case "check.started", "check.finished":
		return fmt.Sprintf("%s · %s", ev.Gate, ev.Check)
	case "gate.started", "gate.finished":
		return ev.Gate
	case "adapter.ready":
		return ev.Adapter + " " + ev.BaseURL
	case "run.failed":
		return ev.Error
	default:
		return ev.Lab
	}
}
