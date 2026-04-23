#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

COMPOSE=(docker compose -f docker-compose.prod.yml)

"${COMPOSE[@]}" down
"${COMPOSE[@]}" up --build -d
"${COMPOSE[@]}" ps
