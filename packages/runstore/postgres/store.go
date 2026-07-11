// Package postgres implements runstore.Store on PostgreSQL via a connection URL.
//
// Use the same style of link as Supabase / Heroku:
//
//	postgres://user:pass@host:5432/dbname?sslmode=require
//
// Env (API): LAB_DATABASE_URL or DATABASE_URL.
package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fastygo/lab/packages/domain"
	"github.com/fastygo/lab/packages/runstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Store is a Postgres-backed runstore.
type Store struct {
	pool *pgxpool.Pool
}

// Open connects using a PostgreSQL connection URI (pgx parses the URL).
func Open(ctx context.Context, databaseURL string) (*Store, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("postgres: empty connection URL")
	}
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse URL: %w", err)
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	s := &Store{pool: pool}
	if err := s.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

// Close releases the pool.
func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

// Migrate applies schema.sql.
func (s *Store) Migrate(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, schemaSQL)
	if err != nil {
		return fmt.Errorf("postgres: migrate: %w", err)
	}
	return nil
}

func (s *Store) CreateRun(ctx context.Context, run *runstore.Run) error {
	if run.ID == "" {
		run.ID = uuid.NewString()
	}
	if run.CreatedAt.IsZero() {
		run.CreatedAt = time.Now().UTC()
	}
	var report []byte
	var err error
	if run.Report != nil {
		report, err = json.Marshal(run.Report)
		if err != nil {
			return err
		}
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO runs (id, lab, status, manifest_json, report_json, error, created_at, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		run.ID, run.Lab, string(run.Status), run.ManifestJSON, nullableJSON(report),
		run.Error, run.CreatedAt, run.StartedAt, run.FinishedAt,
	)
	return err
}

func (s *Store) GetRun(ctx context.Context, id string) (*runstore.Run, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, lab, status, manifest_json, report_json, error, created_at, started_at, finished_at
		FROM runs WHERE id = $1`, id)
	run, err := scanRun(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("run %q not found", id)
		}
		return nil, err
	}
	return run, nil
}

func (s *Store) ListRuns(ctx context.Context, lab string, limit int) ([]*runstore.Run, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows pgx.Rows
	var err error
	if lab == "" {
		rows, err = s.pool.Query(ctx, `
			SELECT id, lab, status, manifest_json, report_json, error, created_at, started_at, finished_at
			FROM runs ORDER BY created_at DESC LIMIT $1`, limit)
	} else {
		rows, err = s.pool.Query(ctx, `
			SELECT id, lab, status, manifest_json, report_json, error, created_at, started_at, finished_at
			FROM runs WHERE lab = $1 ORDER BY created_at DESC LIMIT $2`, lab, limit)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*runstore.Run
	for rows.Next() {
		run, err := scanRun(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, run)
	}
	return out, rows.Err()
}

func (s *Store) UpdateRun(ctx context.Context, run *runstore.Run) error {
	var report []byte
	var err error
	if run.Report != nil {
		report, err = json.Marshal(run.Report)
		if err != nil {
			return err
		}
	}
	tag, err := s.pool.Exec(ctx, `
		UPDATE runs SET lab=$2, status=$3, manifest_json=$4, report_json=$5, error=$6,
		           started_at=$7, finished_at=$8
		WHERE id=$1`,
		run.ID, run.Lab, string(run.Status), run.ManifestJSON, nullableJSON(report),
		run.Error, run.StartedAt, run.FinishedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("run %q not found", run.ID)
	}
	return nil
}

func (s *Store) AppendEvent(ctx context.Context, runID string, ev domain.RunEvent) error {
	if ev.TS.IsZero() {
		ev.TS = time.Now().UTC()
	}
	ev.RunID = runID
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = s.pool.Exec(ctx, `
		INSERT INTO run_events (run_id, type, ts, event_json)
		VALUES ($1, $2, $3, $4)`,
		runID, string(ev.Type), ev.TS, raw,
	)
	return err
}

func (s *Store) ListEvents(ctx context.Context, runID string) ([]domain.RunEvent, error) {
	var exists bool
	if err := s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM runs WHERE id = $1)`, runID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("run %q not found", runID)
	}
	rows, err := s.pool.Query(ctx, `
		SELECT event_json FROM run_events WHERE run_id = $1 ORDER BY id ASC`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.RunEvent
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var ev domain.RunEvent
		if err := json.Unmarshal(raw, &ev); err != nil {
			return nil, err
		}
		out = append(out, ev)
	}
	return out, rows.Err()
}

type scannable interface {
	Scan(dest ...any) error
}

func scanRun(row scannable) (*runstore.Run, error) {
	var (
		run    runstore.Run
		status string
		report []byte
	)
	err := row.Scan(
		&run.ID, &run.Lab, &status, &run.ManifestJSON, &report, &run.Error,
		&run.CreatedAt, &run.StartedAt, &run.FinishedAt,
	)
	if err != nil {
		return nil, err
	}
	run.Status = runstore.RunStatus(status)
	if len(report) > 0 && string(report) != "null" {
		var r domain.Report
		if err := json.Unmarshal(report, &r); err != nil {
			return nil, fmt.Errorf("report json: %w", err)
		}
		run.Report = &r
	}
	return &run, nil
}

func nullableJSON(b []byte) any {
	if len(b) == 0 {
		return nil
	}
	return b
}
