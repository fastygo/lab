package memory

import (
	"context"
	"sync"

	"github.com/fastygo/lab/packages/domain"
)

// Store is an in-memory ArtifactStore.
type Store struct {
	mu      sync.Mutex
	Reports []*domain.Report
}

func New() *Store { return &Store{} }

func (s *Store) SaveReport(_ context.Context, report *domain.Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := *report
	s.Reports = append(s.Reports, &cp)
	return nil
}

// EventSink records RunEvents in memory (tests + local SaaS stubs).
type EventSink struct {
	mu     sync.Mutex
	Events []domain.RunEvent
}

func NewEventSink() *EventSink { return &EventSink{} }

func (s *EventSink) Emit(_ context.Context, ev domain.RunEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Events = append(s.Events, ev)
	return nil
}

// Types returns event type strings in order.
func (s *EventSink) Types() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.Events))
	for i, ev := range s.Events {
		out[i] = string(ev.Type)
	}
	return out
}
