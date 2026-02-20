# Browser Sandbox

Хост для работы с вебом, изображениями и аудио. Основной инструмент — **Jina** (поиск + чтение страниц). Для сложной автоматизации — Playwright + headless Chromium. Для распознавания — OCR и ASR API.

## Системная информация

- ОС: Ubuntu 24.04 (Noble) — образ Microsoft Playwright
- Пользователь: `mantis`
- Домашняя директория: `/home/mantis`
- Shell: `/bin/bash`
- Node.js: предустановлен
- Браузер: Chromium (headless, управляется через Playwright)

## Предустановленное ПО

| Инструмент | Версия | Описание |
|---|---|---|
| **web-search** | — | **Поиск в интернете через DuckDuckGo** (бесплатно, без ключа) |
| **jina-read** | — | **Чтение веб-страниц как Markdown** (бесплатно, без ключа) |
| Playwright | 1.50.0 | Браузерная автоматизация (формы, клики, навигация) |
| Chromium | bundled | Headless-браузер |
| Node.js | 20.x | JavaScript-рантайм |
| npm / npx | bundled | Пакетный менеджер |
| curl, wget | system | HTTP-утилиты |

## Переменные окружения

| Переменная | Описание |
|---|---|
| `OCR_API_URL` | URL сервиса OCR (распознавание текста на изображениях) |
| `ASR_API_URL` | URL сервиса ASR (распознавание речи из аудио) |

## Поиск в интернете

Для поиска информации используй команду `web-search` (DuckDuckGo, бесплатно, без API-ключа):
```bash
web-search <запрос>
```

Пример:
```bash
web-search 'как настроить nginx reverse proxy'
```

Результаты возвращаются в формате Markdown: заголовок, ссылка, сниппет.

## Чтение веб-страниц

Для извлечения содержимого веб-страницы в чистом Markdown используй `jina-read`:
```bash
jina-read <url>
```

Пример:
```bash
jina-read https://docs.python.org/3/tutorial/index.html
```

Jina Reader рендерит JavaScript, поэтому работает с SPA-сайтами (React, Vue и т.д.).

### Когда что использовать

- `web-search` → `jina-read` — **основной флоу**: сначала найди, потом прочитай нужную страницу
- `jina-read` — быстро прочитать текст со страницы, получить статью, документацию
- Playwright — сложная автоматизация: заполнение форм, клики, навигация, скриншоты, перехват запросов

## OCR — распознавание текста на изображениях

Если задана переменная `OCR_API_URL`, можно распознать текст на изображении через curl:

```bash
curl -sS "${OCR_API_URL}/ocr" \
  -F "file=@/home/mantis/image.png" \
  | jq -r '.text'
```

С указанием URL изображения (скачать + OCR):
```bash
curl -sS -L -o /tmp/photo.jpg "https://example.com/photo.jpg"
curl -sS "${OCR_API_URL}/ocr" \
  -F "file=@/tmp/photo.jpg" \
  | jq -r '.text'
```

Поддерживаемые форматы: PNG, JPEG, WebP, BMP, TIFF.

API возвращает JSON:
```json
{"text": "Распознанный текст..."}
```

## ASR — распознавание речи из аудио

Если задана переменная `ASR_API_URL`, можно транскрибировать аудиофайл:

```bash
curl -sS "${ASR_API_URL}/transcribe" \
  -F "file=@/home/mantis/audio.ogg" \
  | jq -r '.text'
```

С URL аудиофайла (скачать + ASR):
```bash
curl -sS -L -o /tmp/voice.ogg "https://example.com/voice.ogg"
curl -sS "${ASR_API_URL}/transcribe" \
  -F "file=@/tmp/voice.ogg" \
  | jq -r '.text'
```

Поддерживаемые форматы: OGG, MP3, WAV, M4A, FLAC.

API возвращает JSON:
```json
{"text": "Распознанный текст из аудио..."}
```

## Скриншот страницы

Для снятия скриншота используй команду `pw-screenshot`:
```bash
pw-screenshot <url> <output.png>
```

Пример:
```bash
pw-screenshot https://example.com screenshot.png
```

Команда ждёт полной загрузки страницы (`networkidle` — отсутствие сетевых запросов 500мс) перед снятием скриншота.

### Дополнительные параметры

С указанием размера вьюпорта:
```bash
pw-screenshot --viewport=1920x1080 https://example.com full.png
```

Полностраничный скриншот (включая прокрутку):
```bash
pw-screenshot --full-page https://example.com fullpage.png
```

### Сохранить страницу как PDF
```bash
npx playwright pdf https://example.com page.pdf
```

### Открыть страницу и получить заголовок
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');
  console.log(await page.title());
  await browser.close();
})();
"
```

## Playwright: работа со страницами (Node.js скрипты)

### Извлечь текст со страницы
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');
  const text = await page.textContent('body');
  console.log(text);
  await browser.close();
})();
"
```

### Извлечь все ссылки со страницы
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');
  const links = await page.$$eval('a[href]', els => els.map(a => ({ text: a.textContent.trim(), href: a.href })));
  console.log(JSON.stringify(links, null, 2));
  await browser.close();
})();
"
```

### Заполнить форму и нажать кнопку
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com/login');
  await page.fill('input[name=username]', 'user');
  await page.fill('input[name=password]', 'pass');
  await page.click('button[type=submit]');
  await page.waitForLoadState('networkidle');
  console.log('URL after login:', page.url());
  await browser.close();
})();
"
```

### Дождаться элемента и получить его содержимое
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');
  await page.waitForSelector('.result');
  const content = await page.textContent('.result');
  console.log(content);
  await browser.close();
})();
"
```

### Перехватить сетевые запросы (мониторинг API)
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  page.on('response', resp => {
    if (resp.url().includes('/api/')) {
      console.log(resp.status(), resp.url());
    }
  });
  await page.goto('https://example.com');
  await page.waitForTimeout(3000);
  await browser.close();
})();
"
```

## Создание переиспользуемых скриптов

Вместо длинных one-liner'ов можно сохранять скрипты в файлы:

```bash
cat > /home/mantis/scrape.js << 'SCRIPT'
const { chromium } = require('playwright');

(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  const url = process.argv[2] || 'https://example.com';
  await page.goto(url);
  console.log(JSON.stringify({
    title: await page.title(),
    url: page.url(),
    text: await page.textContent('body')
  }, null, 2));
  await browser.close();
})();
SCRIPT

node /home/mantis/scrape.js https://example.com
```

## Установка дополнительных npm-пакетов

```bash
npm install -g cheerio    # HTML-парсер
npm install -g puppeteer  # альтернатива Playwright
```

## Ограничения

- Только headless-режим (нет графического дисплея).
- Доступен только Chromium. Firefox и WebKit можно установить: `npx playwright install firefox`.
- Для тяжёлых задач (рендеринг видео, большие PDF) учитывай ограничения памяти контейнера.
- Данные не персистентны — скриншоты и файлы удаляются при перезапуске контейнера.
