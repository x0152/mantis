-- +goose Up

CREATE TABLE skills (
    id            TEXT PRIMARY KEY,
    connection_id TEXT NOT NULL REFERENCES connections(id) ON DELETE CASCADE,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    parameters    JSONB NOT NULL DEFAULT '{}',
    script        TEXT NOT NULL DEFAULT ''
);

CREATE INDEX idx_skills_connection_id ON skills(connection_id);
CREATE UNIQUE INDEX idx_skills_connection_name ON skills(connection_id, name);

-- +goose Down

DROP TABLE IF EXISTS skills;
