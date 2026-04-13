-- +goose Up
ALTER TABLE chat_sessions ADD COLUMN source TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE chat_sessions DROP COLUMN IF EXISTS source;
