-- +goose Up

-- Base sandbox now bundles the former python and db sandboxes: scientific
-- Python 3.12 + DB clients live in the same image. Expand its guard profile
-- to allow Python execution and DB client commands, drop legacy `sudo`
-- capability (the image has no sudo), and turn on codeExec.

UPDATE guard_profiles
SET
  description = 'Alpine Linux + Python 3.12 + DB clients — shell, files, networking, data analysis, database queries',
  capabilities = '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":false,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
  commands = '[
    {"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"find"},
    {"command":"grep"},{"command":"wc"},{"command":"cp"},{"command":"mv"},{"command":"rm"},
    {"command":"mkdir"},{"command":"chmod"},{"command":"chown"},{"command":"tar"},{"command":"gzip"},
    {"command":"unzip"},{"command":"xz"},{"command":"du"},{"command":"df"},{"command":"file"},
    {"command":"tree"},{"command":"stat"},{"command":"ln"},{"command":"touch"},{"command":"awk"},
    {"command":"sed"},{"command":"sort"},{"command":"uniq"},{"command":"cut"},{"command":"tr"},
    {"command":"jq"},{"command":"xargs"},{"command":"echo"},{"command":"printf"},{"command":"tee"},
    {"command":"ps"},{"command":"top"},{"command":"htop"},{"command":"uname"},{"command":"free"},
    {"command":"uptime"},{"command":"env"},{"command":"printenv"},
    {"command":"curl"},{"command":"wget"},{"command":"git"},{"command":"ssh"},{"command":"scp"},
    {"command":"rsync"},{"command":"ping"},{"command":"traceroute"},{"command":"ip"},{"command":"ss"},
    {"command":"dig"},{"command":"host"},{"command":"nslookup"},{"command":"whois"},
    {"command":"python3"},{"command":"ipython"},{"command":"pip"},{"command":"pip3"},
    {"command":"psql"},{"command":"pg_dump"},{"command":"pg_restore"},{"command":"pg_isready"},
    {"command":"mysql"},{"command":"mysqldump"},{"command":"redis-cli"},{"command":"sqlite3"},
    {"command":"apk"}
  ]'
WHERE id = 'base' AND builtin = true;

-- +goose Down

UPDATE guard_profiles
SET
  description = 'Alpine Linux — shell, files, networking, text processing',
  capabilities = '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":true,"codeExec":false,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
  commands = '[{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"find"},{"command":"grep"},{"command":"wc"},{"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"mkdir"},{"command":"chmod"},{"command":"chown"},{"command":"tar"},{"command":"gzip"},{"command":"unzip"},{"command":"du"},{"command":"df"},{"command":"curl"},{"command":"wget"},{"command":"ping"},{"command":"traceroute"},{"command":"ip"},{"command":"ifconfig"},{"command":"netstat"},{"command":"ss"},{"command":"awk"},{"command":"sed"},{"command":"sort"},{"command":"uniq"},{"command":"cut"},{"command":"jq"},{"command":"ps"},{"command":"top"},{"command":"htop"},{"command":"uname"},{"command":"free"},{"command":"uptime"},{"command":"env"},{"command":"printenv"},{"command":"echo"},{"command":"printf"},{"command":"tee"},{"command":"tr"},{"command":"xargs"},{"command":"touch"},{"command":"ln"},{"command":"stat"},{"command":"file"},{"command":"scp"},{"command":"rsync"},{"command":"apk"}]'
WHERE id = 'base' AND builtin = true;
