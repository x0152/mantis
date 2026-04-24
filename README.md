<h1>
  <img src="docs/logo.svg" alt="" width="48" height="48" align="left" />
  &nbsp;Mantis
</h1>

Multi-agent system where an LLM orchestrates a pool of isolated agents, each running on a dedicated SSH sandbox container with specialized tools. Designed for managing large server infrastructure — from quick one-off tasks to complex multi-step workflows. You interact via Telegram or Web UI — the LLM routes tasks to the right agent, commands pass through a guard layer before execution.

> Early development — works end-to-end but expect rough edges.

![demo](docs/demo.gif)

## What it does

- **Chat** — write a message, the LLM picks which server to use and what commands to run
- **Guard** — every command goes through a security layer (profiles with capabilities + command whitelists) before execution
- **Any LLM** — works with any OpenAI-compatible API: cloud or local (Ollama, LM Studio, etc.)
- **Sandboxes** — each server is a Docker container with SSH and pre-installed tools
- **Skills** — reusable SSH scripts exposed as LLM tools with typed parameters and Go template injection
- **Plans** — agentic workflows: visual graph editor (React Flow) with action/decision nodes, branching, retries, clear context, cancel, scheduled execution via cron
  - **Parameters** — plans support typed input parameters (JSON Schema); node prompts use Go templates (`{{.param}}`) for dynamic values
  - **Agent-created plans** — the LLM agent can create multi-step plans from chat using a simple DSL (steps with actions and decisions), including scheduled tasks
- **Presets** — named model configurations (chat model, fallback model, image model) assignable per connection or globally
- **Memory** — long-term memory: remembers facts about you and each server across conversations
- **Notifications** — the agent can send proactive alerts and reports to Telegram via `send_notification`
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
│  │ PostgreSQL │         │ Guard  │──── deny ───▶ x blocked      │
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
| Plans | Visual workflow editor (React Flow), run history, parameters, scheduled execution |
| Skills | Reusable SSH scripts with parameter editor, exposed as agent tools |
| Servers | SSH connection management |
| LLMs & Models | LLM provider connections and model registry |
| Presets | Named model configurations (chat / fallback / image) |
| Channels | Telegram bot configuration |
| Guard Profiles | Security profiles with capability and command whitelists |
| Logs | Session logs with tool call details |

## Quick start

```bash
cp .env.example .env   # fill in AUTH_TOKEN and VITE_LLM_*
./dev.sh
```

Open http://localhost:27173, sign in with your `AUTH_TOKEN`. Done.

## Required env

Drop these into `.env` before anything else:

```bash
AUTH_TOKEN=long-random-string                 # your sign-in token
VITE_LLM_BASE_URL=https://api.openai.com/v1   # or local Ollama / LM Studio
VITE_LLM_API_KEY=sk-...                       # "dummy" for local
VITE_LLM_MODEL=gpt-4o-mini                    # comma-separated for multiple
```

On first start the backend creates a single admin user tied to `AUTH_TOKEN` (change `AUTH_USER_NAME` if you want something other than `admin`). The login endpoint is rate-limited — defaults to 5 failed attempts per 15 minutes per IP; tune with `AUTH_RATE_LIMIT_MAX` / `AUTH_RATE_LIMIT_WINDOW`.

Optional: `VITE_TG_BOT_TOKEN` + `VITE_TG_USER_IDS` (Telegram), `ASR_API_URL` / `OCR_API_URL` / `TTS_API_URL` (speech/OCR services), `MANTIS_BACKEND_PORT` / `MANTIS_FRONTEND_PORT` / `MANTIS_PORT` (host ports). Full list in `.env.example`.

## Generation limits

Caps on how long generation can run and how many tool calls it can make. When a limit kicks in, the assistant message is marked `cancelled` and its content gets a human-readable marker naming the env var to tweak (e.g. `[stopped: supervisor timeout 5m0s exceeded — raise MANTIS_SUPERVISOR_TIMEOUT in .env to increase]`). Partial text and completed tool steps are preserved; unfinished steps get marked `cancelled`. A user-triggered Stop gives `[stopped by user]`.

| Variable | Default | What it caps |
|---|---|---|
| `MANTIS_SUPERVISOR_TIMEOUT` | `5m` | Wall time for one user-message generation by the main agent |
| `MANTIS_SUPERVISOR_MAX_ITERATIONS` | `30` | LLM tool-call rounds the main agent may do per message |
| `MANTIS_SERVER_TIMEOUT` | `5m` | Wall time for one SSH sub-agent call (per `ssh_*` tool invocation) |
| `MANTIS_SERVER_MAX_ITERATIONS` | `30` | LLM tool-call rounds inside one SSH sub-agent call |
| `MANTIS_PLAN_STEP_TIMEOUT` | `10m` | Wall time for a single plan node execution |

Values accept any Go duration (`30s`, `5m`, `1h`). On startup the app logs the active values, e.g. `limits: supervisor=5m0s/30, server=5m0s/30, plan_step=10m0s`. Server-level hits (timeout / iterations) surface as the tool result to the supervisor, so it can read the limit message and adapt instead of failing the whole reply.

## Dev

```bash
./dev.sh
```

Hot reload everywhere — `air` for Go, Vite HMR for the frontend. Frontend on `:27173`, backend on `:27480`, Postgres on `:5432`.

## Prod

```bash
./prod.sh
```

Multi-stage builds, frontend served by nginx, single port `:${MANTIS_PORT:-8080}` exposed, `restart: unless-stopped`.

## Kubernetes

Build and push images once, then:

```bash
helm install mantis ./helm/mantis -n mantis --create-namespace \
  --set ingress.host=mantis.example.com \
  --set app.image.repository=<registry>/mantis       --set app.image.tag=$TAG \
  --set frontend.image.repository=<registry>/mantis-frontend --set frontend.image.tag=$TAG
```

Upgrades:

```bash
helm upgrade mantis ./helm/mantis -n mantis --reuse-values \
  --set app.image.tag=$NEW --set frontend.image.tag=$NEW
```

Chart deploys backend, frontend, Postgres, 6 sandboxes, ingress (`/api` → backend, `/` → frontend), and a `goose` migration job. Build commands and tuning knobs in [`helm/mantis/README.md`](helm/mantis/README.md).

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
