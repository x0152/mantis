-- +goose Up

ALTER TABLE presets
    ADD COLUMN IF NOT EXISTS chat_model_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS summary_model_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS image_model_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS fallback_model_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS temperature DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS system_prompt TEXT NOT NULL DEFAULT '';

ALTER TABLE presets ADD COLUMN IF NOT EXISTS model_id TEXT NOT NULL DEFAULT '';
UPDATE presets
SET chat_model_id = model_id
WHERE COALESCE(chat_model_id, '') = '' AND COALESCE(model_id, '') <> '';

ALTER TABLE presets DROP COLUMN IF EXISTS model_id;

ALTER TABLE channels ADD COLUMN IF NOT EXISTS preset_id TEXT NOT NULL DEFAULT '';
ALTER TABLE connections ADD COLUMN IF NOT EXISTS preset_id TEXT NOT NULL DEFAULT '';
ALTER TABLE connections DROP COLUMN IF EXISTS supervisor_preset_id;

-- +goose Down

ALTER TABLE presets ADD COLUMN IF NOT EXISTS model_id TEXT NOT NULL DEFAULT '';
UPDATE presets
SET model_id = COALESCE(NULLIF(chat_model_id, ''), NULLIF(fallback_model_id, ''), model_id)
WHERE COALESCE(model_id, '') = '';

ALTER TABLE presets DROP COLUMN IF EXISTS chat_model_id;
ALTER TABLE presets DROP COLUMN IF EXISTS summary_model_id;
ALTER TABLE presets DROP COLUMN IF EXISTS image_model_id;
ALTER TABLE presets DROP COLUMN IF EXISTS fallback_model_id;
ALTER TABLE presets DROP COLUMN IF EXISTS temperature;
ALTER TABLE presets DROP COLUMN IF EXISTS system_prompt;

ALTER TABLE connections ADD COLUMN IF NOT EXISTS supervisor_preset_id TEXT NOT NULL DEFAULT '';
