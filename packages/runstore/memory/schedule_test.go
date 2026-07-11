package memory_test

import (
	"context"
	"testing"
	"time"

	"github.com/fastygo/lab/packages/runstore"
	"github.com/fastygo/lab/packages/runstore/memory"
)

func TestScheduleCRUD(t *testing.T) {
	t.Parallel()
	s := memory.New()
	ctx := context.Background()
	next := time.Now().UTC().Add(time.Hour)
	sch := &runstore.Schedule{
		Cron:      "0 * * * *",
		Preset:    "demo",
		Enabled:   true,
		NextRunAt: &next,
	}
	if err := s.CreateSchedule(ctx, sch); err != nil {
		t.Fatal(err)
	}
	if sch.ID == "" {
		t.Fatal("id")
	}
	got, err := s.GetSchedule(ctx, sch.ID)
	if err != nil || got.Preset != "demo" {
		t.Fatalf("get=%v err=%v", got, err)
	}
	got.Enabled = false
	if err := s.UpdateSchedule(ctx, got); err != nil {
		t.Fatal(err)
	}
	list, err := s.ListSchedules(ctx)
	if err != nil || len(list) != 1 || list[0].Enabled {
		t.Fatalf("list=%v err=%v", list, err)
	}
	if err := s.DeleteSchedule(ctx, sch.ID); err != nil {
		t.Fatal(err)
	}
}
