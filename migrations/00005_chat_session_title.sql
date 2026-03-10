-- +goose Up
ALTER TABLE chat_sessions ADD COLUMN title TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE chat_sessions DROP COLUMN title;
