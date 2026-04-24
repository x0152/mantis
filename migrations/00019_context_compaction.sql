-- +goose Up

ALTER TABLE models
    ADD COLUMN IF NOT EXISTS compact_tokens INTEGER NOT NULL DEFAULT 100000;

ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS summary_text TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS summarized_up_to TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS summary_version INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE chat_sessions
    DROP COLUMN IF EXISTS summary_version,
    DROP COLUMN IF EXISTS summarized_up_to,
    DROP COLUMN IF EXISTS summary_text;

ALTER TABLE models
    DROP COLUMN IF EXISTS compact_tokens;
