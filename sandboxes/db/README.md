# Database Sandbox

Host with preinstalled database clients. Connect to any reachable database — PostgreSQL, MySQL, Redis, SQLite.

## System info

- OS: Ubuntu 24.04 (Noble)
- User: `mantis`
- Home directory: `/home/mantis`
- Shell: `/bin/bash`

## Preinstalled clients

| Client | Command | Description |
|---|---|---|
| PostgreSQL | `psql` | PostgreSQL client |
| MySQL / MariaDB | `mysql` | MySQL / MariaDB client |
| Redis | `redis-cli` | Redis client |
| SQLite | `sqlite3` | Embedded file-based DB |
| jq | `jq` | JSON processing (handy for formatting) |
| curl | `curl` | HTTP requests (database REST APIs) |

---

## PostgreSQL (psql)

### Connect
```bash
psql -h <host> -p <port> -U <user> -d <database>
```

Example:
```bash
psql -h postgres -p 5432 -U postgres -d mantis
```

With password via env var:
```bash
PGPASSWORD=mypassword psql -h db.example.com -U admin -d production
```

### Run a SQL query
```bash
psql -h <host> -U <user> -d <database> -c "SELECT * FROM users LIMIT 10;"
```

### CSV output
```bash
psql -h <host> -U <user> -d <database> -c "COPY (SELECT * FROM users) TO STDOUT WITH CSV HEADER;"
```

### List tables
```bash
psql -h <host> -U <user> -d <database> -c "\dt"
```

### Table structure
```bash
psql -h <host> -U <user> -d <database> -c "\d+ tablename"
```

### Database size
```bash
psql -h <host> -U <user> -d <database> -c "SELECT pg_size_pretty(pg_database_size(current_database()));"
```

### Execute a SQL file
```bash
psql -h <host> -U <user> -d <database> -f /home/mantis/query.sql
```

### Dump database
```bash
pg_dump -h <host> -U <user> -d <database> > /home/mantis/dump.sql
```

### Schema-only dump (no data)
```bash
pg_dump -h <host> -U <user> -d <database> --schema-only > /home/mantis/schema.sql
```

---

## MySQL / MariaDB

### Connect
```bash
mysql -h <host> -P <port> -u <user> -p<password> <database>
```

Example:
```bash
mysql -h mysql-server -P 3306 -u root -pMyPassword mydb
```

### Run a query
```bash
mysql -h <host> -u <user> -p<password> <database> -e "SELECT * FROM users LIMIT 10;"
```

### List databases
```bash
mysql -h <host> -u <user> -p<password> -e "SHOW DATABASES;"
```

### List tables
```bash
mysql -h <host> -u <user> -p<password> <database> -e "SHOW TABLES;"
```

### Table structure
```bash
mysql -h <host> -u <user> -p<password> <database> -e "DESCRIBE tablename;"
```

### Dump database
```bash
mysqldump -h <host> -u <user> -p<password> <database> > /home/mantis/dump.sql
```

---

## Redis

### Connect
```bash
redis-cli -h <host> -p <port>
```

With password:
```bash
redis-cli -h <host> -p <port> -a <password>
```

### Run a command
```bash
redis-cli -h <host> PING
redis-cli -h <host> INFO server
redis-cli -h <host> DBSIZE
```

### Get/set a value
```bash
redis-cli -h <host> GET mykey
redis-cli -h <host> SET mykey "value"
```

### Scan keys by pattern
```bash
redis-cli -h <host> --scan --pattern "user:*" | head -20
```

### Monitor commands (real time)
```bash
redis-cli -h <host> MONITOR
```

---

## SQLite

### Create / open a database
```bash
sqlite3 /home/mantis/mydb.sqlite
```

### Run a query
```bash
sqlite3 /home/mantis/mydb.sqlite "SELECT * FROM users;"
```

### Import CSV
```bash
sqlite3 /home/mantis/mydb.sqlite << 'SQL'
.mode csv
.import /home/mantis/data.csv mytable
.schema mytable
SELECT COUNT(*) FROM mytable;
SQL
```

### Export to CSV
```bash
sqlite3 -header -csv /home/mantis/mydb.sqlite "SELECT * FROM users;" > /home/mantis/export.csv
```

---

## Useful patterns

### Save query result to a file
```bash
psql -h postgres -U postgres -d mantis -c "SELECT * FROM users;" > /home/mantis/result.txt
```

### JSON output from PostgreSQL
```bash
psql -h postgres -U postgres -d mantis -t -A -c "SELECT json_agg(t) FROM (SELECT * FROM users LIMIT 5) t;" | jq .
```

### Check host availability
```bash
pg_isready -h postgres -p 5432
mysql -h mysql-server -u root -p -e "SELECT 1;"
redis-cli -h redis-server PING
```

## Limitations

- Data is not persistent — files are deleted when the container restarts.
- For large dumps, consider the container's disk space limits.
- You can connect only to hosts reachable from the Docker network.
