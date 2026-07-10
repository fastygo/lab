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
