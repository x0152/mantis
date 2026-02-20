-- +goose Up

CREATE TABLE configs (
    id   TEXT PRIMARY KEY,
    data JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE llm_connections (
    id       TEXT PRIMARY KEY,
    provider TEXT NOT NULL,
    base_url TEXT NOT NULL,
    api_key  TEXT NOT NULL DEFAULT ''
);

CREATE TABLE models (
    id            TEXT PRIMARY KEY,
    connection_id TEXT NOT NULL,
    name          TEXT NOT NULL,
    thinking_mode TEXT NOT NULL DEFAULT ''
);

CREATE TABLE connections (
    id          TEXT PRIMARY KEY,
    type        TEXT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    model_id    TEXT NOT NULL,
    config      JSONB NOT NULL DEFAULT '{}',
    memories    JSONB NOT NULL DEFAULT '[]'
);

CREATE TABLE channels (
    id               TEXT PRIMARY KEY,
    type             TEXT NOT NULL,
    name             TEXT NOT NULL DEFAULT '',
    token            TEXT NOT NULL DEFAULT '',
    model_id         TEXT NOT NULL DEFAULT '',
    allowed_user_ids JSONB NOT NULL DEFAULT '[]',
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_channels_token ON channels(token) WHERE token <> '';

CREATE TABLE guard_rules (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    pattern       TEXT NOT NULL,
    connection_id TEXT NOT NULL DEFAULT '',
    enabled       BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE cron_jobs (
    id       TEXT PRIMARY KEY,
    name     TEXT NOT NULL,
    schedule TEXT NOT NULL,
    prompt   TEXT NOT NULL DEFAULT '',
    enabled  BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE chat_sessions (
    id         TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE chat_messages (
    id         TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    status     TEXT NOT NULL DEFAULT '',
    steps      JSONB,
    source     TEXT NOT NULL DEFAULT '',
    model_name TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE session_logs (
    id            TEXT PRIMARY KEY,
    connection_id TEXT NOT NULL,
    agent_name    TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'running',
    entries       JSONB NOT NULL DEFAULT '[]',
    prompt        TEXT NOT NULL DEFAULT '',
    message_id    TEXT NOT NULL DEFAULT '',
    step_id       TEXT NOT NULL DEFAULT '',
    model_name    TEXT NOT NULL DEFAULT '',
    started_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at   TIMESTAMPTZ
);
CREATE INDEX idx_session_logs_connection_id ON session_logs(connection_id);
CREATE INDEX idx_session_logs_message_id ON session_logs(message_id);

INSERT INTO configs (id, data) VALUES ('default', '{}');
INSERT INTO channels (id, type, name) VALUES ('chat', 'chat', 'Chat');

-- +goose Down
DROP TABLE session_logs;
DROP TABLE chat_messages;
DROP TABLE chat_sessions;
DROP TABLE cron_jobs;
DROP TABLE guard_rules;
DROP TABLE channels;
DROP TABLE connections;
DROP TABLE models;
DROP TABLE llm_connections;
DROP TABLE configs;
