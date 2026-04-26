# Base Sandbox

General-purpose Linux host based on Ubuntu 24.04.

## System info

- OS: Ubuntu 24.04
- User: `mantis` (passwordless sudo)
- Home directory: `/home/mantis`
- Shell: `/bin/bash`
- SSH: port 22, user `mantis`, password `mantis`

## Preinstalled utilities

### Filesystem
```
ls, cat, head, tail, find, grep, wc    — view and search
cp, mv, rm, mkdir, chmod, chown        — file management
tar, gzip, unzip                       — archiving
du, df                                 — disk usage
tree, file                             — inspect
```

### Network
```
curl, wget         — HTTP requests and downloads
ping, traceroute   — network diagnostics
ip, ifconfig       — network interfaces
netstat, ss        — open connections and ports
dig, whois         — DNS / WHOIS
```

### Text and data
```
awk, sed           — text processing
sort, uniq, cut    — filtering and transformation
jq                 — JSON processing
```

### Processes and system
```
ps, top, htop      — processes
uname -a           — system info
free -h            — memory
uptime             — uptime and load
env, printenv      — environment variables
```

## Installing additional packages

```bash
sudo apt-get update
sudo apt-get install -y <package-name>
```

Examples:
```bash
sudo apt-get install -y python3     # Python 3
sudo apt-get install -y nodejs npm  # Node.js
sudo apt-get install -y git         # Git
```

## Common tasks

### Download a file from the internet
```bash
curl -L -o file.tar.gz https://example.com/file.tar.gz
wget https://example.com/file.tar.gz
```

### Check host availability
```bash
ping -c 3 example.com
curl -I https://example.com
```

### Work with JSON (API responses)
```bash
curl -s https://api.example.com/data | jq '.results[]'
```

### Find files
```bash
find / -name "*.log" -mtime -1     # logs from the last day
find /home -type f -size +10M      # files larger than 10MB
```

## Limitations

- This is a lightweight container; heavy packages (compilers, GUI) are better placed on specialized sandboxes.
- Data is not persistent — everything is reset when the container restarts.
- For browser work use the `browser` sandbox, for media the `ffmpeg` sandbox, for Python data analysis the `python` sandbox.
