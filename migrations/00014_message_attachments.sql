-- +goose Up

ALTER TABLE chat_messages ADD COLUMN attachments JSONB NOT NULL DEFAULT '[]';

-- +goose Down

ALTER TABLE chat_messages DROP COLUMN IF EXISTS attachments;
