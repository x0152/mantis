-- +goose Up

DELETE FROM guard_profiles WHERE builtin = true;

INSERT INTO guard_profiles (id, name, description, builtin, capabilities, commands) VALUES
('base', 'Base Sandbox', 'Alpine Linux — shell, files, networking, text processing', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":true,"codeExec":false,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"find"},{"command":"grep"},{"command":"wc"},{"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"mkdir"},{"command":"chmod"},{"command":"chown"},{"command":"tar"},{"command":"gzip"},{"command":"unzip"},{"command":"du"},{"command":"df"},{"command":"curl"},{"command":"wget"},{"command":"ping"},{"command":"traceroute"},{"command":"ip"},{"command":"ifconfig"},{"command":"netstat"},{"command":"ss"},{"command":"awk"},{"command":"sed"},{"command":"sort"},{"command":"uniq"},{"command":"cut"},{"command":"jq"},{"command":"ps"},{"command":"top"},{"command":"htop"},{"command":"uname"},{"command":"free"},{"command":"uptime"},{"command":"env"},{"command":"printenv"},{"command":"echo"},{"command":"printf"},{"command":"tee"},{"command":"tr"},{"command":"xargs"},{"command":"touch"},{"command":"ln"},{"command":"stat"},{"command":"file"},{"command":"scp"},{"command":"rsync"},{"command":"apk"}]'
),
('browser', 'Browser Sandbox', 'Chromium + Playwright + Jina — web search, reading, screenshots, OCR, ASR', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":false,"sudo":false,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"web-search"},{"command":"jina-read"},{"command":"pw-screenshot"},{"command":"node"},{"command":"npx"},{"command":"npm"},{"command":"curl"},{"command":"wget"},{"command":"jq"},{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"mkdir"},{"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"echo"},{"command":"printf"},{"command":"wc"},{"command":"file"},{"command":"stat"}]'
),
('media', 'Media Sandbox', 'FFmpeg + MediaInfo + ImageMagick — video, audio, image processing', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":false,"sudo":false,"codeExec":false,"download":true,"install":false,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"ffmpeg"},{"command":"ffprobe"},{"command":"mediainfo"},{"command":"convert"},{"command":"mogrify"},{"command":"identify"},{"command":"curl"},{"command":"wget"},{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"mkdir"},{"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"echo"},{"command":"printf"},{"command":"du"},{"command":"df"},{"command":"file"},{"command":"stat"},{"command":"wc"}]'
),
('python', 'Python Sandbox', 'Python 3.12 + numpy, pandas, matplotlib, scikit-learn, requests, beautifulsoup4', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":false,"sudo":false,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"python3"},{"command":"ipython"},{"command":"pip"},{"command":"pip3"},{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"mkdir"},{"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"echo"},{"command":"printf"},{"command":"curl"},{"command":"wget"},{"command":"wc"},{"command":"file"},{"command":"stat"}]'
),
('database', 'Database Sandbox', 'PostgreSQL, MySQL, Redis, SQLite clients + jq', true,
 '{"pipes":true,"redirects":true,"cmdSubst":false,"background":false,"sudo":false,"codeExec":false,"download":false,"install":false,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"psql"},{"command":"pg_dump"},{"command":"pg_restore"},{"command":"pg_isready"},{"command":"mysql"},{"command":"mysqldump"},{"command":"redis-cli"},{"command":"sqlite3"},{"command":"jq"},{"command":"curl"},{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"echo"},{"command":"printf"},{"command":"wc"}]'
),
('unrestricted', 'Unrestricted', 'No restrictions — all commands and capabilities', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":true,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":true,"unrestricted":true}',
 '[]'
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  capabilities = EXCLUDED.capabilities,
  commands = EXCLUDED.commands;

-- +goose Down

DELETE FROM guard_profiles WHERE id IN ('base', 'browser', 'media', 'python', 'database');

INSERT INTO guard_profiles (id, name, description, builtin, capabilities, commands) VALUES
('monitoring', 'Monitoring', 'Read-only monitoring.', true,
 '{"pipes":true,"redirects":false,"cmdSubst":false,"background":false,"sudo":false,"codeExec":false,"download":false,"install":false,"writeFs":false,"networkOut":false,"cron":false,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"ps"},{"command":"top"},{"command":"df"},{"command":"free"}]'
),
('operator', 'Operator', 'Service management.', true,
 '{"pipes":true,"redirects":true,"cmdSubst":false,"background":false,"sudo":true,"codeExec":false,"download":false,"install":false,"writeFs":false,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"systemctl"},{"command":"docker"},{"command":"curl"}]'
),
('database-readonly', 'Database Read-Only', 'Read-only database access.', true,
 '{"pipes":true,"redirects":false,"cmdSubst":false,"background":false,"sudo":false,"codeExec":false,"download":false,"install":false,"writeFs":false,"networkOut":false,"cron":false,"unrestricted":false}',
 '[{"command":"psql","allowedSql":["SELECT","SHOW"]},{"command":"ls"},{"command":"cat"}]'
),
('devops', 'DevOps', 'Full server administration.', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":true,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":true,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"bash"},{"command":"python3"},{"command":"docker"},{"command":"systemctl"}]'
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  capabilities = EXCLUDED.capabilities,
  commands = EXCLUDED.commands;
