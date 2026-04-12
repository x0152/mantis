-- +goose Up

CREATE TABLE settings (
    id               TEXT PRIMARY KEY,
    chat_preset_id   TEXT NOT NULL DEFAULT '',
    server_preset_id TEXT NOT NULL DEFAULT '',
    memory_enabled   BOOLEAN NOT NULL DEFAULT true,
    user_memories    JSONB NOT NULL DEFAULT '[]'
);

INSERT INTO settings (id, chat_preset_id, server_preset_id, memory_enabled, user_memories)
SELECT
    'default',
    COALESCE(data->>'chatPresetId', ''),
    COALESCE(data->>'serverPresetId', ''),
    CASE
        WHEN jsonb_typeof(data->'memoryEnabled') = 'boolean' THEN (data->>'memoryEnabled')::boolean
        ELSE true
    END,
    CASE
        WHEN jsonb_typeof(data->'userMemories') = 'array' THEN data->'userMemories'
        ELSE '[]'::jsonb
    END
FROM configs
WHERE id = 'default'
ON CONFLICT (id) DO UPDATE SET
    chat_preset_id = EXCLUDED.chat_preset_id,
    server_preset_id = EXCLUDED.server_preset_id,
    memory_enabled = EXCLUDED.memory_enabled,
    user_memories = EXCLUDED.user_memories;

INSERT INTO settings (id, chat_preset_id, server_preset_id, memory_enabled, user_memories)
VALUES ('default', '', '', true, '[]'::jsonb)
ON CONFLICT (id) DO NOTHING;

DROP TABLE IF EXISTS configs;

-- +goose Down

CREATE TABLE configs (
    id   TEXT PRIMARY KEY,
    data JSONB NOT NULL DEFAULT '{}'
);

INSERT INTO configs (id, data)
SELECT
    id,
    jsonb_build_object(
        'chatPresetId', chat_preset_id,
        'serverPresetId', server_preset_id,
        'memoryEnabled', memory_enabled,
        'userMemories', user_memories
    )
FROM settings
ON CONFLICT (id) DO UPDATE SET data = EXCLUDED.data;

DROP TABLE IF EXISTS settings;
