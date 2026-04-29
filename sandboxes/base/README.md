# Base Sandbox

General-purpose workhorse host. Combines a small shell environment, network
diagnostic utilities, the most common database clients, and a full Python 3.12
stack with libraries for data analysis, parsing, and visualization. Most data
tasks (download → parse → transform → store) can be done end-to-end here
without hopping between sandboxes.

## System info

- OS: Alpine Linux (`python:3.12-alpine`)
- User: `mantis` (no sudo, unprivileged)
- Home directory: `/home/mantis` (persistent across restarts)
- Shell: `/bin/bash`
- Python: 3.12
- SSH: port 22, user `mantis`, key-only authentication

## Filesystem and process utilities

```
ls, cat, head, tail, find, grep, wc    — view and search
cp, mv, rm, mkdir, chmod, chown        — file management
tar, gzip, xz, unzip                   — archiving
du, df                                 — disk usage
tree, file                             — inspect
ps, top, htop                          — processes
uname -a, free -h, uptime              — system info
env, printenv                          — environment variables
awk, sed, sort, uniq, cut              — text processing
jq                                     — JSON processing
```

## Network utilities

```
curl, wget         — HTTP requests and downloads
ping, traceroute   — network diagnostics
ip, ss             — interfaces and connections
dig, whois         — DNS / WHOIS
ssh, scp           — remote shell, file copy
git                — clone repositories
```

## Database clients

| Client | Command | Description |
|---|---|---|
| PostgreSQL | `psql`, `pg_dump`, `pg_isready` | PostgreSQL client + tooling |
| MySQL / MariaDB | `mysql`, `mysqldump` | MySQL / MariaDB client |
| Redis | `redis-cli` | Redis client |
| SQLite | `sqlite3` | Embedded file-based DB |

### Quick examples

```bash
# Postgres query
PGPASSWORD=secret psql -h db.example.com -U admin -d production \
    -c "SELECT count(*) FROM users;"

# MySQL one-liner
mysql -h mysql -u root -psecret mydb -e "SHOW TABLES;"

# Redis
redis-cli -h redis-server PING
redis-cli -h redis-server --scan --pattern "user:*" | head -20

# SQLite
sqlite3 /home/mantis/mydb.sqlite "SELECT * FROM users LIMIT 10;"
```

## Python

Python 3.12 with the following preinstalled libraries:

| Package | Description |
|---|---|
| ipython | Interactive Python console |
| numpy, pandas | Arrays, dataframes, CSV |
| matplotlib, seaborn | Plots and visualization |
| scipy, sympy | Scientific & symbolic math |
| scikit-learn | Machine learning |
| Pillow | Image processing |
| requests, httpx | HTTP clients |
| beautifulsoup4, lxml | HTML/XML parsing |
| pyyaml, openpyxl | YAML / Excel I/O |
| tabulate, rich, tqdm | Terminal output, progress bars |
| python-dateutil | Date utilities |

### Run a script

```bash
python3 -c "import numpy as np; print(np.array([1,2,3]).mean())"
ipython -c "import pandas as pd; print(pd.DataFrame({'a':[1,2]}))"
```

```bash
cat > /home/mantis/script.py << 'SCRIPT'
import pandas as pd
df = pd.read_csv('/home/mantis/data.csv')
print(df.describe())
SCRIPT
python3 /home/mantis/script.py
```

### Save a plot

```bash
python3 << 'SCRIPT'
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt, numpy as np
x = np.linspace(0, 10, 100)
plt.plot(x, np.sin(x))
plt.savefig('/home/mantis/plot.png', dpi=150)
SCRIPT
```

### Install extra Python packages

The root filesystem is read-only, but `/home/mantis` is writable and
persistent. Use `pip --user`:

```bash
pip install --user openai sqlalchemy fastapi
```

`~/.local/bin` is on `$PATH`, so installed CLI entry points work
immediately.

## Common patterns

### Download and parse JSON

```bash
curl -s https://api.github.com/repos/python/cpython | \
    jq '{stars: .stargazers_count, forks: .forks_count}'
```

### Save a Postgres query as JSON

```bash
PGPASSWORD=secret psql -h db -U admin -d prod -t -A \
    -c "SELECT json_agg(t) FROM (SELECT * FROM orders LIMIT 5) t;" | jq .
```

### Pipe Python into shell

```bash
python3 -c "
import httpx, json
data = httpx.get('https://api.example.com/users').json()
for u in data[:5]: print(u['id'], u['name'])
" | column -t
```

## Limitations

- No GPU — CPU only. Account for runtime on heavy ML tasks.
- The root filesystem is read-only; only `/home/mantis` survives restarts.
  `/tmp`, `/run` and `/var/log` are tmpfs and reset on every boot.
- For browser automation use the `browser` sandbox, for media the `ffmpeg`
  sandbox, for pentest tooling the `netsec` sandbox.
- To send a result to the user, use `ssh_download` + `artifact_send_to_chat`.
