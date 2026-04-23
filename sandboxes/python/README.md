# Python Sandbox

Host for running Python code. Python 3.12 plus preinstalled libraries for data analysis, computation, parsing, and visualization.

## System info

- OS: Debian (python:3.12-slim)
- User: `mantis`
- Home directory: `/home/mantis`
- Shell: `/bin/bash`
- Python: 3.12
- pip: preinstalled

## Preinstalled libraries

| Package | Description |
|---|---|
| ipython | Interactive Python console |
| numpy | Numerical computing, arrays |
| pandas | Dataframes, CSV, data analysis |
| matplotlib | Plots and visualization |
| seaborn | Statistical charts |
| scipy | Scientific computing |
| sympy | Symbolic mathematics |
| scikit-learn | Machine learning |
| Pillow | Image processing |
| requests, httpx | HTTP clients |
| beautifulsoup4, lxml | HTML/XML parsing |
| pyyaml | YAML support |
| tabulate | Table formatting |
| openpyxl | Read/write Excel files |
| python-dateutil | Date utilities |
| tqdm | Progress bars |
| rich | Pretty terminal output |

## Running code

### One-liner
```bash
python3 -c "print(2 ** 100)"
```

### IPython (interactive)
```bash
ipython -c "
import numpy as np
a = np.array([1, 2, 3, 4, 5])
print('mean:', a.mean(), 'std:', a.std())
"
```

### Script from a file
```bash
cat > /home/mantis/script.py << 'SCRIPT'
import pandas as pd

data = {'name': ['Alice', 'Bob'], 'score': [95, 87]}
df = pd.DataFrame(data)
print(df.to_string(index=False))
SCRIPT

python3 /home/mantis/script.py
```

## Common tasks

### CSV analysis
```bash
python3 -c "
import pandas as pd
df = pd.read_csv('data.csv')
print(df.describe())
print(df.head(10).to_string())
"
```

### Plot and save to a file
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
plt.title('Trigonometry')
plt.savefig('/home/mantis/plot.png', dpi=150)
print('Saved to /home/mantis/plot.png')
SCRIPT
```

### HTTP request and JSON parsing
```bash
python3 -c "
import httpx, json
r = httpx.get('https://api.github.com/repos/python/cpython')
data = r.json()
print(f\"Stars: {data['stargazers_count']}, Forks: {data['forks_count']}\")
"
```

### HTML parsing
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

### Computation with sympy
```bash
python3 -c "
from sympy import *
x = Symbol('x')
expr = x**3 - 6*x**2 + 11*x - 6
print('Roots:', solve(expr, x))
print('Derivative:', diff(expr, x))
print('Integral:', integrate(expr, x))
"
```

### Working with images (Pillow)
```bash
python3 -c "
from PIL import Image
img = Image.open('input.png')
print(f'Size: {img.size}, Mode: {img.mode}')
img.thumbnail((800, 800))
img.save('thumb.png')
"
```

### Machine learning (scikit-learn)
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

## Installing additional packages

```bash
pip install <package-name>
```

Examples:
```bash
pip install openai        # OpenAI API
pip install tiktoken      # tokenization
pip install fastapi uvicorn  # web server
pip install sqlalchemy    # ORM
pip install networkx      # graphs
```

## Limitations

- No GPU — CPU only. Account for runtime on heavy ML tasks.
- Data is not persistent — files are deleted when the container restarts.
- Visualization uses the `Agg` backend (no GUI). Save plots to files (PNG, PDF, SVG).
- To send a result to the user, use `ssh_download` + `artifact_send_to_chat`.
