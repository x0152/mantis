# Database Sandbox

Хост с предустановленными клиентами баз данных. Подключайся к любым БД в сети — PostgreSQL, MySQL, Redis, SQLite.

## Системная информация

- ОС: Ubuntu 24.04 (Noble)
- Пользователь: `mantis`
- Домашняя директория: `/home/mantis`
- Shell: `/bin/bash`

## Предустановленные клиенты

| Клиент | Команда | Описание |
|---|---|---|
| PostgreSQL | `psql` | Клиент PostgreSQL |
| MySQL / MariaDB | `mysql` | Клиент MySQL / MariaDB |
| Redis | `redis-cli` | Клиент Redis |
| SQLite | `sqlite3` | Встроенная файловая БД |
| jq | `jq` | Обработка JSON (полезно для форматирования) |
| curl | `curl` | HTTP-запросы (REST API баз данных) |

---

## PostgreSQL (psql)

### Подключение
```bash
psql -h <host> -p <port> -U <user> -d <database>
```

Пример:
```bash
psql -h postgres -p 5432 -U postgres -d mantis
```

С паролем через переменную:
```bash
PGPASSWORD=mypassword psql -h db.example.com -U admin -d production
```

### Выполнить SQL-запрос
```bash
psql -h <host> -U <user> -d <database> -c "SELECT * FROM users LIMIT 10;"
```

### Вывод в формате CSV
```bash
psql -h <host> -U <user> -d <database> -c "COPY (SELECT * FROM users) TO STDOUT WITH CSV HEADER;"
```

### Список таблиц
```bash
psql -h <host> -U <user> -d <database> -c "\dt"
```

### Структура таблицы
```bash
psql -h <host> -U <user> -d <database> -c "\d+ tablename"
```

### Размер базы данных
```bash
psql -h <host> -U <user> -d <database> -c "SELECT pg_size_pretty(pg_database_size(current_database()));"
```

### Выполнить SQL-файл
```bash
psql -h <host> -U <user> -d <database> -f /home/mantis/query.sql
```

### Дамп базы
```bash
pg_dump -h <host> -U <user> -d <database> > /home/mantis/dump.sql
```

### Дамп только структуры (без данных)
```bash
pg_dump -h <host> -U <user> -d <database> --schema-only > /home/mantis/schema.sql
```

---

## MySQL / MariaDB

### Подключение
```bash
mysql -h <host> -P <port> -u <user> -p<password> <database>
```

Пример:
```bash
mysql -h mysql-server -P 3306 -u root -pMyPassword mydb
```

### Выполнить запрос
```bash
mysql -h <host> -u <user> -p<password> <database> -e "SELECT * FROM users LIMIT 10;"
```

### Список баз данных
```bash
mysql -h <host> -u <user> -p<password> -e "SHOW DATABASES;"
```

### Список таблиц
```bash
mysql -h <host> -u <user> -p<password> <database> -e "SHOW TABLES;"
```

### Структура таблицы
```bash
mysql -h <host> -u <user> -p<password> <database> -e "DESCRIBE tablename;"
```

### Дамп базы
```bash
mysqldump -h <host> -u <user> -p<password> <database> > /home/mantis/dump.sql
```

---

## Redis

### Подключение
```bash
redis-cli -h <host> -p <port>
```

С паролем:
```bash
redis-cli -h <host> -p <port> -a <password>
```

### Выполнить команду
```bash
redis-cli -h <host> PING
redis-cli -h <host> INFO server
redis-cli -h <host> DBSIZE
```

### Получить/установить значение
```bash
redis-cli -h <host> GET mykey
redis-cli -h <host> SET mykey "value"
```

### Поиск ключей по паттерну
```bash
redis-cli -h <host> --scan --pattern "user:*" | head -20
```

### Мониторинг команд (в реальном времени)
```bash
redis-cli -h <host> MONITOR
```

---

## SQLite

### Создать / открыть базу
```bash
sqlite3 /home/mantis/mydb.sqlite
```

### Выполнить запрос
```bash
sqlite3 /home/mantis/mydb.sqlite "SELECT * FROM users;"
```

### Импорт CSV
```bash
sqlite3 /home/mantis/mydb.sqlite << 'SQL'
.mode csv
.import /home/mantis/data.csv mytable
.schema mytable
SELECT COUNT(*) FROM mytable;
SQL
```

### Экспорт в CSV
```bash
sqlite3 -header -csv /home/mantis/mydb.sqlite "SELECT * FROM users;" > /home/mantis/export.csv
```

---

## Полезные паттерны

### Запрос с сохранением результата
```bash
psql -h postgres -U postgres -d mantis -c "SELECT * FROM users;" > /home/mantis/result.txt
```

### JSON-вывод из PostgreSQL
```bash
psql -h postgres -U postgres -d mantis -t -A -c "SELECT json_agg(t) FROM (SELECT * FROM users LIMIT 5) t;" | jq .
```

### Проверка доступности хоста
```bash
pg_isready -h postgres -p 5432
mysql -h mysql-server -u root -p -e "SELECT 1;"
redis-cli -h redis-server PING
```

## Ограничения

- Данные не персистентны — файлы удаляются при перезапуске контейнера.
- Для дампов больших баз учитывай ограничения дискового пространства контейнера.
- Подключение возможно только к хостам, доступным из Docker-сети.
