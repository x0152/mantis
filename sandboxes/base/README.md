# Base Sandbox

Базовый Linux-хост общего назначения на Alpine Linux.

## Системная информация

- ОС: Alpine Linux (linuxserver/openssh-server)
- Пользователь: `mantis`
- Домашняя директория: `/config` (linuxserver convention)
- Shell: `/bin/bash`

## Предустановленные утилиты

### Файловая система
```
ls, cat, head, tail, find, grep, wc    — просмотр и поиск
cp, mv, rm, mkdir, chmod, chown        — управление файлами
tar, gzip, unzip                       — архивация
du, df                                  — использование диска
```

### Сеть
```
curl, wget         — HTTP-запросы и загрузка файлов
ping, traceroute   — диагностика сети
ip, ifconfig       — сетевые интерфейсы
netstat, ss        — открытые соединения и порты
```

### Текст и данные
```
awk, sed           — обработка текста
sort, uniq, cut    — фильтрация и трансформация
jq                 — работа с JSON (если установлен)
```

### Процессы и система
```
ps, top, htop      — процессы
uname -a           — информация о системе
free -h            — память
uptime             — аптайм и нагрузка
env, printenv      — переменные окружения
```

## Как установить дополнительные пакеты

```bash
sudo apk update
sudo apk add <package-name>
```

Примеры:
```bash
sudo apk add jq          # JSON-процессор
sudo apk add git         # Git
sudo apk add python3     # Python 3
sudo apk add nodejs npm  # Node.js
```

## Типичные задачи

### Загрузить файл из интернета
```bash
curl -L -o file.tar.gz https://example.com/file.tar.gz
wget https://example.com/file.tar.gz
```

### Проверить доступность хоста
```bash
ping -c 3 example.com
curl -I https://example.com
```

### Работа с JSON (API-ответами)
```bash
curl -s https://api.example.com/data | jq '.results[]'
```

### Найти файлы
```bash
find / -name "*.log" -mtime -1     # логи за последний день
find /home -type f -size +10M      # файлы больше 10MB
```

## Ограничения

- Это легковесный контейнер, тяжёлые пакеты (компиляторы, GUI) лучше ставить на специализированные хосты.
- Данные не персистентны — при перезапуске контейнера всё сбрасывается.
- Для работы с браузером используй `browser-sandbox`, для медиа — `ffmpeg-sandbox`.
