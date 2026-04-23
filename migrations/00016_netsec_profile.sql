-- +goose Up

INSERT INTO guard_profiles (id, name, description, builtin, capabilities, commands) VALUES
('netsec', 'Netsec Sandbox',
 'Pentest/netsec toolkit — nmap, dig, whois, curl, openssl, nikto, sqlmap, ffuf, gobuster, whatweb, hashcat, john + net-* wrappers with hard timeouts. Use ONLY on targets you have explicit permission to test.',
 true,
 '{"pipes":true,"redirects":true,"cmdSubst":true,"background":false,"sudo":false,"codeExec":false,"download":true,"install":false,"writeFs":true,"networkOut":true,"cron":false,"unrestricted":false}',
 '[
   {"command":"net-port"},{"command":"net-http"},{"command":"net-headers"},{"command":"net-tls"},
   {"command":"net-dns"},{"command":"net-whois"},{"command":"net-dir"},{"command":"net-subs"},
   {"command":"net-whatweb"},{"command":"net-vuln"},{"command":"net-hash-id"},{"command":"net-hash-crack"},
   {"command":"net-banner"},{"command":"net-ping"},
   {"command":"nmap"},{"command":"dig"},{"command":"host"},{"command":"nslookup"},{"command":"whois"},
   {"command":"curl"},{"command":"wget"},{"command":"nc"},{"command":"ncat"},
   {"command":"openssl"},{"command":"testssl"},
   {"command":"nikto"},{"command":"sqlmap"},
   {"command":"ffuf"},{"command":"gobuster"},{"command":"dirb"},{"command":"wfuzz"},
   {"command":"whatweb"},{"command":"dnsrecon"},
   {"command":"hashcat"},{"command":"john"},{"command":"hashid"},
   {"command":"ping"},{"command":"traceroute"},{"command":"mtr"},
   {"command":"jq"},{"command":"timeout"},
   {"command":"ls"},{"command":"cat"},{"command":"head"},{"command":"tail"},{"command":"grep"},
   {"command":"find"},{"command":"wc"},{"command":"file"},{"command":"stat"},
   {"command":"sort"},{"command":"uniq"},{"command":"awk"},{"command":"sed"},{"command":"cut"},{"command":"tr"},
   {"command":"cp"},{"command":"mv"},{"command":"rm"},{"command":"mkdir"},{"command":"touch"},
   {"command":"echo"},{"command":"printf"},{"command":"tee"},{"command":"xargs"}
 ]'
)
ON CONFLICT (id) DO UPDATE SET
  name = EXCLUDED.name,
  description = EXCLUDED.description,
  capabilities = EXCLUDED.capabilities,
  commands = EXCLUDED.commands;

-- +goose Down

DELETE FROM guard_profiles WHERE id = 'netsec';
