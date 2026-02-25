-- +goose Up

ALTER TABLE connections ADD COLUMN memory_enabled BOOLEAN NOT NULL DEFAULT true;

-- +goose Down

ALTER TABLE connections DROP COLUMN memory_enabled;
