#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/.local/spx/worker.pid"

if [[ ! -f "$PID_FILE" ]]; then
  printf 'Background worker is stopped.\n'
  exit 0
fi

pid="$(<"$PID_FILE")"
if [[ "$pid" =~ ^[0-9]+$ ]] && kill -0 "$pid" 2>/dev/null; then
  printf 'Background worker is running (PID %s).\n' "$pid"
else
  printf 'Background worker has a stale PID file.\n'
  exit 1
fi
