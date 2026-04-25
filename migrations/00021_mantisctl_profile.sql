-- +goose Up

INSERT INTO guard_profiles (id, name, description, builtin, capabilities, commands) VALUES
('mantisctl', 'Mantisctl Sandbox',
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

-- +goose Down

DELETE FROM guard_profiles WHERE id = 'mantisctl';
