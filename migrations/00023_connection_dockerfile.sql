-- +goose Up

ALTER TABLE connections ADD COLUMN IF NOT EXISTS dockerfile TEXT NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE connections DROP COLUMN IF EXISTS dockerfile;
