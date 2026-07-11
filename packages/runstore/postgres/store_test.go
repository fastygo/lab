package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/runstore"
	"github.com/fastygo/lab/packages/runstore/postgres"
)

func TestPostgresStore(t *testing.T) {
	url := os.Getenv("LAB_DATABASE_URL")
	if url == "" {
		url = os.Getenv("DATABASE_URL")
	}
	if url == "" {
		t.Skip("set LAB_DATABASE_URL or DATABASE_URL to run Postgres store tests")
	}
	ctx := context.Background()
	s, err := postgres.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	run := &runstore.Run{
		Lab:          "demo",
		Status:       runstore.StatusQueued,
		ManifestJSON: []byte("apiVersion: lab.fastygo.dev/v1\n"),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.CreateRun(ctx, run); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetRun(ctx, run.ID)
	if err != nil || got.Lab != "demo" {
		t.Fatalf("get=%v err=%v", got, err)
	}

	ev := domain.RunEvent{Type: domain.EventRunStarted, Lab: "demo", TS: time.Now().UTC()}
	if err := s.AppendEvent(ctx, run.ID, ev); err != nil {
		t.Fatal(err)
	}
	events, err := s.ListEvents(ctx, run.ID)
	if err != nil || len(events) != 1 {
		t.Fatalf("events=%v err=%v", events, err)
	}

	now := time.Now().UTC()
	got.Status = runstore.StatusPass
	got.FinishedAt = &now
	got.Report = &domain.Report{Status: domain.StatusPass, Summary: domain.Summary{Total: 1}}
	if err := s.UpdateRun(ctx, got); err != nil {
		t.Fatal(err)
	}
	again, err := s.GetRun(ctx, run.ID)
	if err != nil || again.Status != runstore.StatusPass || again.Report == nil {
		t.Fatalf("updated=%v err=%v", again, err)
	}
	list, err := s.ListRuns(ctx, "demo", 10)
	if err != nil || len(list) < 1 {
		t.Fatalf("list=%v err=%v", list, err)
	}
}
