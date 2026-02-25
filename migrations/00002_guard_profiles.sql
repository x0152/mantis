-- +goose Up

CREATE TABLE guard_profiles (
    id           TEXT PRIMARY KEY,
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    builtin      BOOLEAN NOT NULL DEFAULT false,
    capabilities JSONB NOT NULL DEFAULT '{}',
    commands     JSONB NOT NULL DEFAULT '[]'
);

ALTER TABLE connections ADD COLUMN profile_ids JSONB NOT NULL DEFAULT '[]';

INSERT INTO guard_profiles (id, name, description, builtin, capabilities, commands) VALUES
('monitoring', 'Monitoring', 'Read-only monitoring. No writes, no installs, no code execution.', true,
 '{"pipes":true,"redirects":false,"cmdSubst":false,"background":false,"sudo":false,"codeExec":false,"download":false,"install":false,"writeFs":false,"networkOut":false,"cron":false,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"wc"},{"command":"stat"},{"command":"file"},{"command":"df"},{"command":"du"},{"command":"free"},{"command":"ps"},{"command":"top"},{"command":"htop"},{"command":"uptime"},{"command":"uname"},{"command":"whoami"},{"command":"hostname"},{"command":"id"},{"command":"date"},{"command":"env"},{"command":"printenv"},{"command":"which"},{"command":"journalctl"},{"command":"systemctl","allowedArgs":["status","show","is-active","list-units","list-timers"]},{"command":"docker","allowedArgs":["ps","logs","inspect","stats","images"]},{"command":"ip"},{"command":"ss"},{"command":"netstat"},{"command":"lsof"}]'
),
('operator', 'Operator', 'Service management. Can restart services, view logs, run diagnostics.', true,
 '{"pipes":true,"redirects":true,"cmdSubst":false,"background":false,"sudo":true,"codeExec":false,"download":false,"install":false,"writeFs":false,"networkOut":true,"cron":false,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"wc"},{"command":"stat"},{"command":"file"},{"command":"df"},{"command":"du"},{"command":"free"},{"command":"ps"},{"command":"top"},{"command":"htop"},{"command":"uptime"},{"command":"uname"},{"command":"whoami"},{"command":"hostname"},{"command":"id"},{"command":"date"},{"command":"env"},{"command":"printenv"},{"command":"which"},{"command":"journalctl"},{"command":"systemctl"},{"command":"service"},{"command":"docker"},{"command":"nginx","allowedArgs":["-t","-s"]},{"command":"ip"},{"command":"ss"},{"command":"netstat"},{"command":"lsof"},{"command":"curl"},{"command":"ping"},{"command":"dig"},{"command":"traceroute"},{"command":"nslookup"}]'
),
('database-readonly', 'Database Read-Only', 'Monitoring plus read-only database access.', true,
 '{"pipes":true,"redirects":false,"cmdSubst":false,"background":false,"sudo":false,"codeExec":false,"download":false,"install":false,"writeFs":false,"networkOut":false,"cron":false,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"wc"},{"command":"stat"},{"command":"file"},{"command":"df"},{"command":"du"},{"command":"free"},{"command":"ps"},{"command":"top"},{"command":"uptime"},{"command":"uname"},{"command":"whoami"},{"command":"hostname"},{"command":"id"},{"command":"date"},{"command":"journalctl"},{"command":"systemctl","allowedArgs":["status","show","is-active"]},{"command":"psql","allowedSql":["SELECT","SHOW","EXPLAIN","\\dt","\\l","\\d+","\\di","\\dn","\\du","\\df"]},{"command":"mysql","allowedSql":["SELECT","SHOW","DESCRIBE","EXPLAIN","USE"]},{"command":"redis-cli","allowedArgs":["GET","MGET","KEYS","INFO","DBSIZE","TTL","TYPE","SCAN","HGETALL","LRANGE","SCARD","SMEMBERS"]}]'
),
('devops', 'DevOps', 'Full server administration. Can install packages, write files, execute code.', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":true,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":true,"unrestricted":false}',
 '[{"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},{"command":"find"},{"command":"wc"},{"command":"stat"},{"command":"file"},{"command":"df"},{"command":"du"},{"command":"free"},{"command":"ps"},{"command":"top"},{"command":"htop"},{"command":"uptime"},{"command":"uname"},{"command":"whoami"},{"command":"hostname"},{"command":"id"},{"command":"date"},{"command":"env"},{"command":"printenv"},{"command":"which"},{"command":"journalctl"},{"command":"systemctl"},{"command":"service"},{"command":"docker"},{"command":"nginx"},{"command":"ip"},{"command":"ss"},{"command":"netstat"},{"command":"lsof"},{"command":"curl"},{"command":"wget"},{"command":"ping"},{"command":"dig"},{"command":"traceroute"},{"command":"nslookup"},{"command":"apt"},{"command":"apt-get"},{"command":"yum"},{"command":"dnf"},{"command":"pip"},{"command":"pip3"},{"command":"npm"},{"command":"cp"},{"command":"mv"},{"command":"mkdir"},{"command":"touch"},{"command":"tee"},{"command":"chmod"},{"command":"chown"},{"command":"ln"},{"command":"rm"},{"command":"tar"},{"command":"gzip"},{"command":"gunzip"},{"command":"zip"},{"command":"unzip"},{"command":"sed"},{"command":"awk"},{"command":"bash"},{"command":"sh"},{"command":"python"},{"command":"python3"},{"command":"node"},{"command":"crontab"},{"command":"scp"},{"command":"rsync"},{"command":"git"},{"command":"make"},{"command":"vi"},{"command":"nano"},{"command":"echo"},{"command":"printf"},{"command":"sort"},{"command":"uniq"},{"command":"cut"},{"command":"tr"},{"command":"xargs"},{"command":"kill"},{"command":"pkill"},{"command":"nohup"},{"command":"screen"},{"command":"tmux"},{"command":"psql"},{"command":"mysql"},{"command":"redis-cli"},{"command":"pg_dump"},{"command":"mysqldump"}]'
),
('unrestricted', 'Unrestricted', 'No restrictions. All commands and capabilities allowed.', true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":true,"sudo":true,"codeExec":true,"download":true,"install":true,"writeFs":true,"networkOut":true,"cron":true,"unrestricted":true}',
 '[]'
);

DROP TABLE guard_rules;

-- +goose Down

ALTER TABLE connections DROP COLUMN profile_ids;
DROP TABLE guard_profiles;

CREATE TABLE guard_rules (
    id            TEXT PRIMARY KEY,
    name          TEXT NOT NULL,
    description   TEXT NOT NULL DEFAULT '',
    pattern       TEXT NOT NULL,
    connection_id TEXT NOT NULL DEFAULT '',
    enabled       BOOLEAN NOT NULL DEFAULT true
);
