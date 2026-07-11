-- Cycle F runstore schema (Postgres / Supabase-compatible).
-- Applied by packages/runstore/postgres.Migrate.

CREATE TABLE IF NOT EXISTS runs (
    id              UUID PRIMARY KEY,
    lab             TEXT NOT NULL,
    status          TEXT NOT NULL,
    manifest_json   BYTEA NOT NULL DEFAULT ''::bytea,
    report_json     JSONB,
    error           TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL,
    started_at      TIMESTAMPTZ,
    finished_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS runs_lab_created_idx ON runs (lab, created_at DESC);
CREATE INDEX IF NOT EXISTS runs_created_idx ON runs (created_at DESC);

CREATE TABLE IF NOT EXISTS run_events (
    id          BIGSERIAL PRIMARY KEY,
    run_id      UUID NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    type        TEXT NOT NULL,
    ts          TIMESTAMPTZ NOT NULL,
    event_json  JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS run_events_run_id_idx ON run_events (run_id, id);

CREATE TABLE IF NOT EXISTS artifacts (
    id            UUID PRIMARY KEY,
    run_id        UUID NOT NULL REFERENCES runs (id) ON DELETE CASCADE,
    kind          TEXT NOT NULL,
    name          TEXT NOT NULL DEFAULT '',
    content_type  TEXT NOT NULL DEFAULT 'application/octet-stream',
    bytes         BYTEA,
    uri           TEXT NOT NULL DEFAULT '',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS artifacts_run_id_idx ON artifacts (run_id, created_at);

CREATE TABLE IF NOT EXISTS schedules (
    id           UUID PRIMARY KEY,
    cron         TEXT NOT NULL,
    preset       TEXT NOT NULL,
    lab          TEXT NOT NULL DEFAULT '',
    enabled      BOOLEAN NOT NULL DEFAULT true,
    theme_zip    TEXT NOT NULL DEFAULT '',
    base_url     TEXT NOT NULL DEFAULT '',
    root         TEXT NOT NULL DEFAULT '',
    last_run_at  TIMESTAMPTZ,
    next_run_at  TIMESTAMPTZ,
    created_at   TIMESTAMPTZ NOT NULL,
    updated_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS schedules_enabled_next_idx ON schedules (enabled, next_run_at);
