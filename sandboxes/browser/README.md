# Browser Sandbox

Host for working with the web, images, and audio. The main tool is **Jina** (search + page reading). For complex automation — Playwright + headless Chromium. For recognition — OCR and ASR APIs.

## System info

- OS: Ubuntu 24.04 (Noble) — Microsoft Playwright image
- User: `mantis`
- Home directory: `/home/mantis`
- Shell: `/bin/bash`
- Node.js: preinstalled
- Browser: Chromium (headless, driven via Playwright)

## Preinstalled software

| Tool | Version | Description |
|---|---|---|
| **web-search** | — | **Web search via DuckDuckGo** (free, no key required) |
| **jina-read** | — | **Read web pages as Markdown** (free, no key required) |
| Playwright | 1.50.0 | Browser automation (forms, clicks, navigation) |
| Chromium | bundled | Headless browser |
| Node.js | 20.x | JavaScript runtime |
| npm / npx | bundled | Package manager |
| curl, wget | system | HTTP utilities |

## Environment variables

| Variable | Description |
|---|---|
| `OCR_API_URL` | URL of the OCR service (recognize text in images) |
| `ASR_API_URL` | URL of the ASR service (speech-to-text from audio) |

## Web search

To search for information use `web-search` (DuckDuckGo, free, no API key):
```bash
web-search <query>
```

Example:
```bash
web-search 'how to set up nginx reverse proxy'
```

Results are returned as Markdown: title, link, snippet.

## Reading web pages

To extract page content as clean Markdown use `jina-read`:
```bash
jina-read <url>
```

Example:
```bash
jina-read https://docs.python.org/3/tutorial/index.html
```

Jina Reader renders JavaScript, so it works with SPA sites (React, Vue, etc.).

### When to use what

- `web-search` → `jina-read` — **the main flow**: first find, then read the relevant page.
- `jina-read` — quickly read text from a page, grab an article or documentation.
- Playwright — complex automation: filling forms, clicking, navigation, screenshots, request interception.

## Execution rules

Use these rules to avoid unnecessary failures and retries:

1. **Default to text tools first**
   - For extraction tasks ("get title/url", "read article", "summarize page"), use `jina-read` or `curl` + parser.
   - Do **not** start with Playwright unless the task needs interaction or visual rendering.

2. **Use Playwright only when required**
   - Required cases: login flows, clicking, submitting forms, waiting for post-click state, screenshots, JS-only content that text tools cannot read.

3. **Prefer script files over long one-liners**
   - If the command is longer than a few lines, write it to `/home/mantis/<name>.js` and run `node /home/mantis/<name>.js`.
   - This avoids shell escaping issues and makes debugging easier.

4. **Never put `$$eval` inside double-quoted shell strings**
   - In `node -e "..."`, shell expands `$$` to PID and breaks Playwright code.
   - Either:
     - use a script file (`cat << 'SCRIPT' ...`), or
     - wrap with single quotes and avoid shell interpolation.

5. **Avoid runtime installs unless explicitly requested**
   - Core dependencies and Playwright browser binaries are preinstalled in this image.
   - Do not run `npm install` / `npx playwright install` unless the user asked for extra packages.

## OCR — text recognition in images

If `OCR_API_URL` is set, you can recognize text in an image via curl:

```bash
curl -sS "${OCR_API_URL}/ocr" \
  -F "file=@/home/mantis/image.png" \
  | jq -r '.text'
```

With an image URL (download + OCR):
```bash
curl -sS -L -o /tmp/photo.jpg "https://example.com/photo.jpg"
curl -sS "${OCR_API_URL}/ocr" \
  -F "file=@/tmp/photo.jpg" \
  | jq -r '.text'
```

Supported formats: PNG, JPEG, WebP, BMP, TIFF.

The API returns JSON:
```json
{"text": "Recognized text..."}
```

## ASR — speech-to-text from audio

If `ASR_API_URL` is set, you can transcribe an audio file:

```bash
curl -sS "${ASR_API_URL}/transcribe" \
  -F "file=@/home/mantis/audio.ogg" \
  | jq -r '.text'
```

With an audio URL (download + ASR):
```bash
curl -sS -L -o /tmp/voice.ogg "https://example.com/voice.ogg"
curl -sS "${ASR_API_URL}/transcribe" \
  -F "file=@/tmp/voice.ogg" \
  | jq -r '.text'
```

Supported formats: OGG, MP3, WAV, M4A, FLAC.

The API returns JSON:
```json
{"text": "Recognized text from audio..."}
```

## Page screenshot

To take a screenshot use `pw-screenshot`:
```bash
pw-screenshot <url> <output.png>
```

Example:
```bash
pw-screenshot https://example.com screenshot.png
```

The command waits for the page to fully load (`networkidle` — no network requests for 500ms) before taking the screenshot.

### Extra options

With a specific viewport size:
```bash
pw-screenshot --viewport=1920x1080 https://example.com full.png
```

Full-page screenshot (including scroll):
```bash
pw-screenshot --full-page https://example.com fullpage.png
```

### Save page as PDF
```bash
npx playwright pdf https://example.com page.pdf
```

### Open a page and get its title
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

## Playwright: working with pages (Node.js scripts)

`playwright` and Chromium are preinstalled and available without extra setup.

### Extract text from a page
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

### Extract all links from a page
```bash
node -e "
const { chromium } = require('playwright');
(async () => {
  const browser = await chromium.launch();
  const page = await browser.newPage();
  await page.goto('https://example.com');
  const links = await page.\$\$eval('a[href]', els => els.map(a => ({ text: a.textContent.trim(), href: a.href })));
  console.log(JSON.stringify(links, null, 2));
  await browser.close();
})();
"
```

### Fill a form and click a button
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

### Wait for an element and read its content
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

### Intercept network requests (API monitoring)
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

## Reusable scripts

Instead of long one-liners you can store scripts in files:

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

## Installing additional npm packages

```bash
npm install -g cheerio    # HTML parser
npm install -g puppeteer  # alternative to Playwright
```

In most cases, this step is unnecessary because `playwright` and `cheerio` are already installed in the image.

## Troubleshooting

### `Cannot find module 'playwright'`

`NODE_PATH` is preconfigured in this sandbox. If you still see this error, run:

```bash
node -e "console.log(require.resolve('playwright'))"
```

If resolution fails, the image is likely outdated and should be rebuilt.

### `Executable doesn't exist ... chromium ...`

Chromium is preinstalled in `/ms-playwright`. If this appears, verify env:

```bash
echo "$PLAYWRIGHT_BROWSERS_PATH"
```

If empty, set it for the current shell:

```bash
export PLAYWRIGHT_BROWSERS_PATH=/ms-playwright
```

## Limitations

- Headless mode only (no graphical display).
- Only Chromium is available. Firefox and WebKit can be installed: `npx playwright install firefox`.
- For heavy tasks (video rendering, large PDFs) mind the container's memory limits.
- Data is not persistent — screenshots and files are deleted when the container restarts.
