package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/fastygo/lab/packages/runstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *Store) CreateSchedule(ctx context.Context, sch *runstore.Schedule) error {
	if sch.ID == "" {
		sch.ID = uuid.NewString()
	}
	now := time.Now().UTC()
	if sch.CreatedAt.IsZero() {
		sch.CreatedAt = now
	}
	sch.UpdatedAt = now
	_, err := s.pool.Exec(ctx, `
		INSERT INTO schedules (
			id, cron, preset, lab, enabled, theme_zip, base_url, root,
			last_run_at, next_run_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		sch.ID, sch.Cron, sch.Preset, sch.Lab, sch.Enabled, sch.ThemeZip, sch.BaseURL, sch.Root,
		sch.LastRunAt, sch.NextRunAt, sch.CreatedAt, sch.UpdatedAt,
	)
	return err
}

func (s *Store) GetSchedule(ctx context.Context, id string) (*runstore.Schedule, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, cron, preset, lab, enabled, theme_zip, base_url, root,
		       last_run_at, next_run_at, created_at, updated_at
		FROM schedules WHERE id = $1`, id)
	sch, err := scanSchedule(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("schedule %q not found", id)
		}
		return nil, err
	}
	return sch, nil
}

func (s *Store) ListSchedules(ctx context.Context) ([]*runstore.Schedule, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, cron, preset, lab, enabled, theme_zip, base_url, root,
		       last_run_at, next_run_at, created_at, updated_at
		FROM schedules ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*runstore.Schedule
	for rows.Next() {
		sch, err := scanSchedule(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, sch)
	}
	return out, rows.Err()
}

func (s *Store) UpdateSchedule(ctx context.Context, sch *runstore.Schedule) error {
	sch.UpdatedAt = time.Now().UTC()
	tag, err := s.pool.Exec(ctx, `
		UPDATE schedules SET cron=$2, preset=$3, lab=$4, enabled=$5, theme_zip=$6, base_url=$7, root=$8,
		                last_run_at=$9, next_run_at=$10, updated_at=$11
		WHERE id=$1`,
		sch.ID, sch.Cron, sch.Preset, sch.Lab, sch.Enabled, sch.ThemeZip, sch.BaseURL, sch.Root,
		sch.LastRunAt, sch.NextRunAt, sch.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("schedule %q not found", sch.ID)
	}
	return nil
}

func (s *Store) DeleteSchedule(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM schedules WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("schedule %q not found", id)
	}
	return nil
}

func scanSchedule(row scannable) (*runstore.Schedule, error) {
	var sch runstore.Schedule
	err := row.Scan(
		&sch.ID, &sch.Cron, &sch.Preset, &sch.Lab, &sch.Enabled,
		&sch.ThemeZip, &sch.BaseURL, &sch.Root,
		&sch.LastRunAt, &sch.NextRunAt, &sch.CreatedAt, &sch.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sch, nil
}
