-- +goose Up

CREATE TABLE users (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL DEFAULT '',
    api_key_hash TEXT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_users_api_key_hash ON users(api_key_hash);

-- +goose Down

DROP TABLE users;
