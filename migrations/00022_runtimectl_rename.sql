-- +goose Up

UPDATE guard_profiles SET id = 'runtimectl', name = 'Runtimectl Sandbox' WHERE id = 'mantisctl';

INSERT INTO guard_profiles (id, name, description, builtin, capabilities, commands) VALUES
('runtimectl', 'Runtimectl Sandbox',
 'Runtime controller. Lets the agent build Docker images and run new sandboxes through mantisctl (a thin curl/jq wrapper around Mantis /api/runtime). Shell tools are limited to writing Dockerfiles, driving mantisctl, and grepping its output.',
 true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":false,"sudo":false,"codeExec":false,"download":false,"install":false,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[
   {"command":"mantisctl"},
   {"command":"ssh"},{"command":"scp"},
   {"command":"curl"},{"command":"wget"},
   {"command":"jq"},{"command":"yq"},
   {"command":"cat"},{"command":"ls"},{"command":"head"},{"command":"tail"},
   {"command":"grep"},{"command":"find"},{"command":"wc"},{"command":"file"},
   {"command":"awk"},{"command":"sed"},{"command":"sort"},{"command":"uniq"},{"command":"cut"},{"command":"tr"},
   {"command":"mkdir"},{"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"touch"},
   {"command":"echo"},{"command":"printf"},{"command":"tee"},{"command":"xargs"},
   {"command":"timeout"},
   {"command":"less"},{"command":"more"}
 ]'
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  capabilities = EXCLUDED.capabilities,
  commands = EXCLUDED.commands;

UPDATE connections SET name = 'runtimectl', profile_ids = '["runtimectl"]'::jsonb
WHERE name = 'mantisctl' AND type = 'ssh';

UPDATE connections
SET config = jsonb_set(config, '{host}', '"runtimectl-sandbox"'::jsonb)
WHERE name = 'runtimectl'
  AND type = 'ssh'
  AND config->>'host' = 'mantisctl-sandbox';

-- +goose Down

UPDATE connections
SET config = jsonb_set(config, '{host}', '"mantisctl-sandbox"'::jsonb)
WHERE name = 'runtimectl'
  AND type = 'ssh'
  AND config->>'host' = 'runtimectl-sandbox';

UPDATE connections SET name = 'mantisctl', profile_ids = '["mantisctl"]'::jsonb
WHERE name = 'runtimectl' AND type = 'ssh';

DELETE FROM guard_profiles WHERE id = 'runtimectl';
