# Base Sandbox

General-purpose Linux host based on Alpine Linux.

## System info

- OS: Alpine Linux (linuxserver/openssh-server)
- User: `mantis`
- Home directory: `/config` (linuxserver convention)
- Shell: `/bin/bash`

## Preinstalled utilities

### Filesystem
```
ls, cat, head, tail, find, grep, wc    — view and search
cp, mv, rm, mkdir, chmod, chown        — file management
tar, gzip, unzip                       — archiving
du, df                                  — disk usage
```

### Network
```
curl, wget         — HTTP requests and downloads
ping, traceroute   — network diagnostics
ip, ifconfig       — network interfaces
netstat, ss        — open connections and ports
```

### Text and data
```
awk, sed           — text processing
sort, uniq, cut    — filtering and transformation
jq                 — JSON processing (if installed)
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
sudo apk update
sudo apk add <package-name>
```

Examples:
```bash
sudo apk add jq          # JSON processor
sudo apk add git         # Git
sudo apk add python3     # Python 3
sudo apk add nodejs npm  # Node.js
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

- This is a lightweight container; heavy packages (compilers, GUI) are better placed on specialized hosts.
- Data is not persistent — everything is reset when the container restarts.
- For browser work use `browser-sandbox`, for media use `ffmpeg-sandbox`.
