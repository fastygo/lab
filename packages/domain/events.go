package domain

import "time"

// RunEventType is a progress event for SaaS dashboards / workers (Cycle F).
type RunEventType string

const (
	EventRunStarted     RunEventType = "run.started"
	EventAdapterReady   RunEventType = "adapter.ready"
	EventGateStarted    RunEventType = "gate.started"
	EventCheckStarted   RunEventType = "check.started"
	EventCheckFinished  RunEventType = "check.finished"
	EventGateFinished   RunEventType = "gate.finished"
	EventRunFinished    RunEventType = "run.finished"
	EventRunFailed      RunEventType = "run.failed"
)

// RunEvent is one lifecycle notification during orchestrator.Run.
// CLI may ignore; API/worker persists for timelines and SSE.
type RunEvent struct {
	Type      RunEventType      `json:"type"`
	TS        time.Time         `json:"ts"`
	Lab       string            `json:"lab,omitempty"`
	RunID     string            `json:"runId,omitempty"`
	Gate      string            `json:"gate,omitempty"`
	Check     string            `json:"check,omitempty"`
	Runner    string            `json:"runner,omitempty"`
	Adapter   string            `json:"adapter,omitempty"`
	BaseURL   string            `json:"baseUrl,omitempty"`
	Status    ReportStatus      `json:"status,omitempty"`
	Error     string            `json:"error,omitempty"`
	Payload   map[string]string `json:"payload,omitempty"`
}
