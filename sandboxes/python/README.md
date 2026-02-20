# Python Sandbox

Хост для выполнения Python-кода. Python 3.12 + предустановленные библиотеки для анализа данных, вычислений, парсинга и визуализации.

## Системная информация

- ОС: Debian (python:3.12-slim)
- Пользователь: `mantis`
- Домашняя директория: `/home/mantis`
- Shell: `/bin/bash`
- Python: 3.12
- pip: предустановлен

## Предустановленные библиотеки

| Пакет | Описание |
|---|---|
| ipython | Интерактивная Python-консоль |
| numpy | Числовые вычисления, массивы |
| pandas | Таблицы, CSV, анализ данных |
| matplotlib | Графики и визуализация |
| seaborn | Статистические графики |
| scipy | Научные вычисления |
| sympy | Символьная математика |
| scikit-learn | Машинное обучение |
| Pillow | Обработка изображений |
| requests, httpx | HTTP-запросы |
| beautifulsoup4, lxml | Парсинг HTML/XML |
| pyyaml | Работа с YAML |
| tabulate | Форматирование таблиц |
| openpyxl | Чтение/запись Excel-файлов |
| python-dateutil | Работа с датами |
| tqdm | Прогресс-бары |
| rich | Красивый вывод в терминал |

## Выполнение кода

### Однострочник
```bash
python3 -c "print(2 ** 100)"
```

### IPython (интерактивный)
```bash
ipython -c "
import numpy as np
a = np.array([1, 2, 3, 4, 5])
print('mean:', a.mean(), 'std:', a.std())
"
```

### Скрипт из файла
```bash
cat > /home/mantis/script.py << 'SCRIPT'
import pandas as pd

data = {'name': ['Alice', 'Bob'], 'score': [95, 87]}
df = pd.DataFrame(data)
print(df.to_string(index=False))
SCRIPT

python3 /home/mantis/script.py
```

## Типичные задачи

### Анализ CSV
```bash
python3 -c "
import pandas as pd
df = pd.read_csv('data.csv')
print(df.describe())
print(df.head(10).to_string())
"
```

### Построить график и сохранить в файл
```bash
python3 << 'SCRIPT'
import matplotlib
matplotlib.use('Agg')
import matplotlib.pyplot as plt
import numpy as np

x = np.linspace(0, 10, 100)
plt.figure(figsize=(10, 6))
plt.plot(x, np.sin(x), label='sin(x)')
plt.plot(x, np.cos(x), label='cos(x)')
plt.legend()
plt.title('Тригонометрия')
plt.savefig('/home/mantis/plot.png', dpi=150)
print('Saved to /home/mantis/plot.png')
SCRIPT
```

### HTTP-запрос и парсинг JSON
```bash
python3 -c "
import httpx, json
r = httpx.get('https://api.github.com/repos/python/cpython')
data = r.json()
print(f\"Stars: {data['stargazers_count']}, Forks: {data['forks_count']}\")
"
```

### Парсинг HTML
```bash
python3 -c "
import httpx
from bs4 import BeautifulSoup
r = httpx.get('https://example.com')
soup = BeautifulSoup(r.text, 'lxml')
print(soup.title.string)
print(soup.get_text()[:500])
"
```

### Вычисления с sympy
```bash
python3 -c "
from sympy import *
x = Symbol('x')
expr = x**3 - 6*x**2 + 11*x - 6
print('Корни:', solve(expr, x))
print('Производная:', diff(expr, x))
print('Интеграл:', integrate(expr, x))
"
```

### Работа с изображениями (Pillow)
```bash
python3 -c "
from PIL import Image
img = Image.open('input.png')
print(f'Size: {img.size}, Mode: {img.mode}')
img.thumbnail((800, 800))
img.save('thumb.png')
"
```

### Машинное обучение (scikit-learn)
```bash
python3 << 'SCRIPT'
from sklearn.datasets import load_iris
from sklearn.model_selection import train_test_split
from sklearn.ensemble import RandomForestClassifier
from sklearn.metrics import accuracy_score

X, y = load_iris(return_X_y=True)
X_train, X_test, y_train, y_test = train_test_split(X, y, test_size=0.3)
clf = RandomForestClassifier(n_estimators=100)
clf.fit(X_train, y_train)
print(f'Accuracy: {accuracy_score(y_test, clf.predict(X_test)):.2%}')
SCRIPT
```

## Установка дополнительных пакетов

```bash
pip install <package-name>
```

Примеры:
```bash
pip install openai        # OpenAI API
pip install tiktoken      # токенизация
pip install fastapi uvicorn  # веб-сервер
pip install sqlalchemy    # ORM
pip install networkx      # графы
```

## Ограничения

- Нет GPU — только CPU-вычисления. Для тяжёлого ML учитывай время работы.
- Данные не персистентны — файлы удаляются при перезапуске контейнера.
- Для визуализации используется `Agg` backend (без GUI). Графики сохраняются в файлы (PNG, PDF, SVG).
- Для отправки результата пользователю используй ssh_download + artifact_send_to_chat.
