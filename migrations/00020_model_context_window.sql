-- +goose Up

ALTER TABLE models
    ADD COLUMN IF NOT EXISTS context_window INTEGER NOT NULL DEFAULT 128000,
    ADD COLUMN IF NOT EXISTS reserve_tokens INTEGER NOT NULL DEFAULT 20000;

ALTER TABLE chat_sessions
    ADD COLUMN IF NOT EXISTS last_context_tokens INTEGER NOT NULL DEFAULT 0;

ALTER TABLE chat_messages
    ADD COLUMN IF NOT EXISTS prompt_tokens INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS completion_tokens INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE chat_messages
    DROP COLUMN IF EXISTS completion_tokens,
    DROP COLUMN IF EXISTS prompt_tokens;

ALTER TABLE chat_sessions
    DROP COLUMN IF EXISTS last_context_tokens;

ALTER TABLE models
    DROP COLUMN IF EXISTS reserve_tokens,
    DROP COLUMN IF EXISTS context_window;
