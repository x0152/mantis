-- +goose Up

CREATE TABLE presets (
    id                TEXT PRIMARY KEY,
    name              TEXT NOT NULL,
    chat_model_id     TEXT NOT NULL DEFAULT '',
    summary_model_id  TEXT NOT NULL DEFAULT '',
    image_model_id    TEXT NOT NULL DEFAULT '',
    fallback_model_id TEXT NOT NULL DEFAULT '',
    temperature       DOUBLE PRECISION,
    system_prompt     TEXT NOT NULL DEFAULT ''
);

ALTER TABLE channels ADD COLUMN preset_id TEXT NOT NULL DEFAULT '';
ALTER TABLE connections ADD COLUMN preset_id TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE connections DROP COLUMN preset_id;
ALTER TABLE channels DROP COLUMN preset_id;
DROP TABLE presets;
