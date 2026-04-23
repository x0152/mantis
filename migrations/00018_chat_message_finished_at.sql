-- +goose Up

ALTER TABLE chat_messages ADD COLUMN finished_at TIMESTAMPTZ;

-- +goose Down

ALTER TABLE chat_messages DROP COLUMN IF EXISTS finished_at;
