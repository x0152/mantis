-- +goose Up

CREATE TABLE plan_runs (
    id          TEXT PRIMARY KEY,
    plan_id     TEXT NOT NULL REFERENCES plans(id) ON DELETE CASCADE,
    status      TEXT NOT NULL DEFAULT 'running',
    trigger     TEXT NOT NULL DEFAULT 'manual',
    steps       JSONB NOT NULL DEFAULT '[]',
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    finished_at TIMESTAMPTZ
);

CREATE INDEX idx_plan_runs_plan_id ON plan_runs(plan_id);
CREATE INDEX idx_plan_runs_status ON plan_runs(status);

-- +goose Down

DROP TABLE IF EXISTS plan_runs;
