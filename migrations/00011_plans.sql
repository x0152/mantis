-- +goose Up

CREATE TABLE plans (
    id          TEXT PRIMARY KEY,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    schedule    TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT false,
    graph       JSONB NOT NULL DEFAULT '{"nodes":[],"edges":[]}'
);

-- +goose Down

DROP TABLE IF EXISTS plans;
