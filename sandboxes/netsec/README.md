# Netsec Sandbox

A host for network diagnostics and light pentesting: ports, HTTP, TLS, DNS, directory brute-force, web stack fingerprinting, vulnerability scanning, and hash identification/cracking.

**The core principle is speed.** Every `net-*` wrapper is executed through `timeout`, so no check can hang for more than a couple of minutes. For long-running scans, use the raw tools directly and set your own timeouts.

## System info

- OS: Ubuntu 24.04 (Noble)
- User: `mantis` / `mantis`
- Home directory: `/home/mantis`
- Shell: `/bin/bash`

## Pre-installed tooling

| Tool | Purpose |
|---|---|
| **nmap** | port scanning, service detection |
| **dig / host / nslookup** | DNS lookups |
| **whois** | domain and IP registration info |
| **curl / wget** | HTTP clients |
| **netcat (nc)** | TCP/UDP cat, banner grabbing |
| **openssl** | TLS diagnostics, s_client |
| **testssl.sh** (`testssl`) | deep TLS audit |
| **nikto** | quick web vulnerability scan |
| **sqlmap** | SQL injection testing |
| **ffuf / gobuster / dirb / wfuzz** | fuzzing paths, parameters, vhosts |
| **whatweb** | web stack fingerprinting |
| **dnsrecon** | DNS enumeration |
| **hashcat / john / hashid** | hash identification and cracking |
| **ping / traceroute / mtr** | network diagnostics |

## Quick `net-*` wrappers

All wrappers enforce hard timeouts so nothing hangs.

| Command | Purpose |
|---|---|
| `net-port <host> [ports]` | fast nmap port scan (top100 / common / web / all / list) |
| `net-http <url>` | status, headers, redirects, timings |
| `net-headers <url>` | audit of security headers (HSTS, CSP, XFO, …) |
| `net-tls <host[:port]>` | certificate, protocols, ciphers |
| `net-dns <domain> [resolver]` | A/AAAA/MX/NS/TXT/SOA/CAA/SRV + PTR |
| `net-whois <domain\|ip>` | whois with a timeout |
| `net-dir <url> [wordlist]` | directory brute-force via ffuf |
| `net-subs <domain>` | subdomain brute-force |
| `net-whatweb <url>` | web stack fingerprinting |
| `net-vuln <url>` | quick Nikto scan (limited tuning) |
| `net-hash-id <hash>` | identify hash type |
| `net-hash-crack <hash> [format] [wordlist]` | crack hash via Python + passlib |
| `net-banner <host> <port>` | grab banner from a port |
| `net-ping <host>` | ping 4 packets, 10s timeout |

Running any command with no arguments prints its usage.

## Typical scenarios

### Check which ports are open
```bash
net-port example.com            # top-100 (default, ~10s)
net-port 10.0.0.1 common        # 16 most common ports
net-port target.lab web         # web ports only
net-port target.lab 22,80,443   # specific ports
```

### Inspect HTTP surface
```bash
net-http https://example.com
net-headers https://example.com
net-whatweb https://example.com
```

### Check TLS certificate and protocols
```bash
net-tls example.com
net-tls mail.example.com:993
# deep audit (slower):
testssl --fast example.com
```

### DNS and infra
```bash
net-dns example.com
net-dns example.com 8.8.8.8     # use a specific resolver
net-subs example.com            # subdomain DNS brute-force
net-whois example.com
```

### Find hidden paths on a site
```bash
net-dir https://example.com                 # small wordlist ~50 paths
net-dir https://example.com dirb-common     # ~4600 paths, up to 60s
net-dir https://example.com /home/mantis/custom.txt
```

### Identify a hash type and try to crack it
```bash
net-hash-id 5f4dcc3b5aa765d61d8327deb882cf99
net-hash-crack 5f4dcc3b5aa765d61d8327deb882cf99 raw-md5
net-hash-crack '$2a$10$...'    bcrypt john     # bcrypt with john's wordlist
net-hash-crack '$1$xy$...'     md5crypt /home/mantis/mydict.txt
```

Formats supported by the wrapper: `raw-md5`, `raw-sha1`, `raw-sha256`, `raw-sha512`, `nt`, `md5crypt`, `sha256crypt`, `sha512crypt`, `bcrypt`. Implementation is Python + passlib (works reliably without OpenCL/GPU on short wordlists). For heavy brute-forcing use `hashcat` / `john` directly.

### Grab a banner from a port
```bash
net-banner example.com 22       # SSH banner
net-banner mail.example.com 25  # SMTP banner
```

### Web vulnerability scan
```bash
net-vuln https://example.com    # Nikto, limited, up to 2 minutes
```

## Wordlists

Small wordlists optimised for fast checks live in `/usr/share/wordlists/mantis/`:

- `passwords-top.txt` — ~50 most common passwords
- `paths-top.txt` — ~50 web paths (admin, api, .env, backup, …)
- `subdomains-top.txt` — ~50 subdomains (www, mail, api, dev, …)

Larger wordlists shipped with system packages:

- `/usr/share/dirb/wordlists/common.txt` — ~4600 paths
- `/usr/share/john/password.lst` — ~3500 passwords
- `/usr/share/wfuzz/wordlist/` — various wfuzz wordlists
- `/usr/share/dnsrecon/` — wordlists for DNS enumeration

Additional wordlists can be pulled in:
```bash
git clone --depth=1 https://github.com/danielmiessler/SecLists /home/mantis/SecLists
```

## Using raw utilities

If the wrappers are not enough, call tools directly — but **always** wrap them in `timeout`:

```bash
timeout 60 nmap -sV -sC -T4 --top-ports 100 target
timeout 90 sqlmap -u "https://example.com/?id=1" --batch --level=1 --risk=1 --random-agent
timeout 60 ffuf -u "https://example.com/FUZZ" -w /usr/share/dirb/wordlists/common.txt -mc 200,301,302,403 -t 40
timeout 120 gobuster dir -u https://example.com -w /usr/share/dirb/wordlists/common.txt -t 40 -q
timeout 120 hashcat -m 0 -a 0 hash.txt /usr/share/wordlists/mantis/passwords-top.txt --runtime=100
```

## Ethical constraints

Scan and brute-force **only** hosts you have explicit permission to test (your own services, CTF environments, bug-bounty scope). Active probing of third-party systems without the owner's consent is illegal in most jurisdictions.

## Environment limits

- Data is not persistent — everything is wiped when the container restarts.
- No GPU available → `hashcat` runs on CPU only and may fail to find OpenCL devices in arm64 containers. For fast dictionary attacks on standard formats use `net-hash-crack` (pure Python implementation).
- Some UDP scans and ICMP probes can give imprecise results inside a Docker network.
- Hard timeouts are designed for lightweight checks; for anything long-running, start it in `screen`/`tmux` or set your own `timeout`.
