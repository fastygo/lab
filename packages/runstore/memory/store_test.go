package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/runstore"
	"github.com/fastygo/lab/packages/runstore/memory"
)

func TestRunStoreMemory(t *testing.T) {
	t.Parallel()
	s := memory.New()
	ctx := context.Background()
	run := &runstore.Run{
		Lab:       "quality",
		Status:    runstore.StatusQueued,
		CreatedAt: time.Now().UTC(),
	}
	if err := s.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	if run.ID == "" {
		t.Fatal("expected id")
	}
	got, err := s.GetRun(ctx, run.ID)
	if err != nil || got.Lab != "quality" {
		t.Fatalf("get=%v err=%v", got, err)
	}
	ev := domain.RunEvent{Type: domain.EventRunStarted, Lab: "quality", TS: time.Now().UTC()}
	if err := s.AppendEvent(ctx, run.ID, ev); err != nil {
		t.Fatal(err)
	}
	events, err := s.ListEvents(ctx, run.ID)
	if err != nil || len(events) != 1 {
		t.Fatalf("events=%v err=%v", events, err)
	}
}
