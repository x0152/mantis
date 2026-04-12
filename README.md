# Mantis

Multi-agent system where an LLM orchestrates a pool of isolated agents, each running on a dedicated SSH sandbox container with specialized tools. Designed for managing large server infrastructure — from quick one-off tasks to complex multi-step workflows. You interact via Telegram or Web UI — the LLM routes tasks to the right agent, commands pass through a guard layer before execution.

> Early development — works end-to-end but expect rough edges.

![demo](docs/demo.gif)

## What it does

- **Chat** — write a message, the LLM picks which server to use and what commands to run
- **Guard** — every command goes through a security layer (profiles with capabilities + command whitelists) before execution
- **Any LLM** — works with any OpenAI-compatible API: cloud or local (Ollama, LM Studio, etc.)
- **Sandboxes** — each server is a Docker container with SSH and pre-installed tools
- **Skills** — reusable SSH scripts exposed as LLM tools with typed parameters and Go template injection
- **Plans** — agentic workflows: visual graph editor (React Flow) with action/decision nodes, branching, retries, clear context, cancel, scheduled execution
- **Cron** — scheduled single-prompt jobs with Telegram delivery ("send me BTC price every morning")
- **Presets** — named model configurations (chat model, fallback model, image model) assignable per connection or globally
- **Memory** — long-term memory: remembers facts about you and each server across conversations
- **Telegram** — bot with voice messages, files, model switching
- **ASR / OCR / TTS** — optional speech-to-text, OCR, text-to-speech integrations

## Architecture

```
                                                ┌──────────────────┐
┌───────────┐  ┌───────────┐                    │  LLM provider    │
│ Telegram  │  │ Web Chat  │                    │  (OpenAI / local)│
└─────┬─────┘  └─────┬─────┘                    └────────┬─────────┘
      │               │                                  │ API
      ▼               ▼                                  │
┌────────────────────────────────────────────────────────┼────────┐
│  Mantis                          docker-compose / k8s  │        │
│                                                        │        │
│  ┌─────────────┐   ┌──────────────────┐          ┌─────┴──────┐ │
│  │  Web Panel  │   │   Agent Loop     │◀────────▶│ LLM client │ │
│  │   (React)   │   │                  │          └────────────┘ │
│  └─────────────┘   └────────┬─────────┘                         │
│                          tool calls                             │
│  ┌────────────┐         ┌───┴────┐                              │
│  │ PostgreSQL │         │ Guard  │──── deny ───▶ ✕ blocked      │
│  └────────────┘         └───┬────┘                              │
│                           allow                                 │
│                    ┌────────┼────────┐                           │
│                    ▼        ▼        ▼                           │
│               ┌────────┬────────┬────────┬────────┐             │
│               │ agent  │ agent  │ agent  │ agent  │  ...        │
│               └───┬────┘───┬────┘───┬────┘───┬────┘             │
└───────────────────┼────────┼────────┼────────┼──────────────────┘
                    │        │        │        │ SSH
                    ▼        ▼        ▼        ▼
              ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐
              │  base  │ │browser │ │ ffmpeg │ │ python │ │   db   │
              │  :2222 │ │ :2223  │ │ :2224  │ │ :2225  │ │  :2226 │
              └────────┘ └────────┘ └────────┘ └────────┘ └────────┘
                    isolated SSH sandboxes with pre-installed tools
```

## Web Panel

| Page | Description |
|------|-------------|
| Chat | Conversations with the agent, session management |
| Plans | Visual workflow editor (React Flow), run history with step-by-step logs |
| Skills | Reusable SSH scripts with parameter editor, exposed as agent tools |
| Cron Jobs | Scheduled tasks with Telegram delivery |
| Servers | SSH connection management |
| LLMs & Models | LLM provider connections and model registry |
| Presets | Named model configurations (chat / fallback / image) |
| Channels | Telegram bot configuration |
| Guard Profiles | Security profiles with capability and command whitelists |
| Logs | Session logs with tool call details |

## Quick start

```bash
docker compose up --build
```

Starts Postgres, runs migrations, API on `:27480`, frontend on `:27173`, and 5 SSH sandboxes (internal network only, not exposed to host ports).

Go to http://localhost:27173 — first time it'll ask for your LLM provider URL and API key. Sets up the model, sandbox connections, and optionally a Telegram bot. After that just start chatting.

## Dev setup

```bash
# postgres + sandboxes
docker compose up postgres sandbox browser-sandbox ffmpeg-sandbox python-sandbox db-sandbox -d

# migrations
go install github.com/pressly/goose/v3/cmd/goose@latest
goose -dir migrations postgres "postgres://postgres:postgres@localhost:5432/mantis?sslmode=disable" up

# backend (hot reload)
go install github.com/air-verse/air@latest
air

# frontend
cd frontend && pnpm install && pnpm dev
```

Backend on http://localhost:8080, frontend on http://localhost:5173 (proxies `/api` to backend).

## Environment variables

See `.env.example` for defaults. Key variables:

| Variable | Description |
|----------|-------------|
| `DATABASE_URL` | Postgres connection string |
| `PORT` | Backend port (default `8080`) |
| `CRON_DELIVERY_CHANNEL` | Where cron results go (`telegram` or empty) |
| `ASR_API_URL` | Speech-to-text service URL (optional) |
| `OCR_API_URL` | OCR service URL (optional) |
| `TTS_API_URL` | Text-to-speech service URL (optional) |

First-run wizard variables (used by `docker-compose.yml` for auto-setup):

| Variable | Description |
|----------|-------------|
| `VITE_LLM_BASE_URL` | LLM provider base URL |
| `VITE_LLM_API_KEY` | LLM provider API key |
| `VITE_LLM_MODEL` | Default model name |
| `VITE_TG_BOT_TOKEN` | Telegram bot token (optional) |
| `VITE_TG_USER_IDS` | Allowed Telegram user IDs (optional) |

## ASR, OCR & TTS (optional)

| Service | Env var | Repo |
|---------|---------|------|
| Speech-to-text | `ASR_API_URL` | [russian-asr](https://github.com/x0152/russian-asr) |
| OCR | `OCR_API_URL` | [easy-ocr-api](https://github.com/x0152/easy-ocr-api) |
| Text-to-speech | `TTS_API_URL` | [cosyvoice-tts-api](https://github.com/x0152/cosyvoice-tts-api) |

```bash
docker run -p 8016:8016 ghcr.io/x0152/russian-asr        # --gpus all for CUDA
docker run -p 8017:8017 ghcr.io/x0152/easy-ocr-api
docker run -p 8020:8020 ghcr.io/x0152/cosyvoice-tts-api
```

Set the URLs in `.env` (see `.env.example`).

## License

MIT
