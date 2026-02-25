# Mantis

Multi-agent system where an LLM orchestrates a pool of isolated agents, each running on a dedicated SSH sandbox container with specialized tools. Designed for managing large server infrastructure - from quick one-off tasks to ongoing operations like monitoring, deployments, and scheduled jobs. You interact via Telegram or Web UI - the LLM routes tasks to the right agent, commands pass through a guard layer before execution.

> Early development — works end-to-end but expect rough edges.

![demo](docs/demo.gif)

## What it does

- You write a message, the LLM picks which server to use and what commands to run
- Every command goes through a guard (security profiles with capabilities + command whitelists) before execution
- Works with any OpenAI-compatible API — cloud or local (Ollama, LM Studio, etc.)
- Each server is a Docker container with SSH and pre-installed tools
- Long-term memory — remembers facts about you and each server across conversations
- Web panel for managing everything: connections, models, guard profiles, cron jobs
- Telegram bot with voice messages, files, model switching
- Cron jobs — tell it "send me BTC price every morning" and it will
- Optional speech-to-text, OCR, text-to-speech

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

## Quick start

```bash
docker compose up --build
```

Starts postgres, runs migrations, API on `:8080`, frontend on `:5173`, and 5 SSH sandboxes.

Go to http://localhost:5173 — first time it'll ask for your LLM provider URL and API key. Sets up the model, sandbox connections, and optionally a Telegram bot. After that just start chatting.

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
