package memory

import (
	"context"
	"fmt"
	"sync"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/runstore"
	"github.com/google/uuid"
)

// Store is an in-memory runstore.Store for F0/F1 stubs.
type Store struct {
	mu     sync.Mutex
	runs   map[string]*runstore.Run
	events map[string][]domain.RunEvent
	order  []string
}

func New() *Store {
	return &Store{
		runs:   map[string]*runstore.Run{},
		events: map[string][]domain.RunEvent{},
	}
}

func (s *Store) CreateRun(_ context.Context, run *runstore.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	cp := *run
	s.runs[cp.ID] = &cp
	s.order = append(s.order, cp.ID)
	return nil
}

func (s *Store) GetRun(_ context.Context, id string) (*runstore.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.runs[id]
	if !ok {
		return nil, fmt.Errorf("run %q not found", id)
	}
	cp := *r
	return &cp, nil
}

func (s *Store) ListRuns(_ context.Context, lab string, limit int) ([]*runstore.Run, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if limit <= 0 {
		limit = 50
	}
	var out []*runstore.Run
	for i := len(s.order) - 1; i >= 0 && len(out) < limit; i-- {
		r := s.runs[s.order[i]]
		if lab != "" && r.Lab != lab {
			continue
		}
		cp := *r
		out = append(out, &cp)
	}
	return out, nil
}

func (s *Store) UpdateRun(_ context.Context, run *runstore.Run) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[run.ID]; !ok {
		return fmt.Errorf("run %q not found", run.ID)
	}
	cp := *run
	s.runs[run.ID] = &cp
	return nil
}

func (s *Store) AppendEvent(_ context.Context, runID string, ev domain.RunEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[runID]; !ok {
		return fmt.Errorf("run %q not found", runID)
	}
	s.events[runID] = append(s.events[runID], ev)
	return nil
}

func (s *Store) ListEvents(_ context.Context, runID string) ([]domain.RunEvent, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.runs[runID]; !ok {
		return nil, fmt.Errorf("run %q not found", runID)
	}
	src := s.events[runID]
	out := make([]domain.RunEvent, len(src))
	copy(out, src)
	return out, nil
}
