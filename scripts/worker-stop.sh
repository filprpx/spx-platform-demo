#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_FILE="${ROOT_DIR}/.local/spx/worker.pid"

if [[ ! -f "$PID_FILE" ]]; then
  printf 'Background worker is not running.\n'
  exit 0
fi

pid="$(<"$PID_FILE")"
if [[ ! "$pid" =~ ^[0-9]+$ ]]; then
  rm -f "$PID_FILE"
  printf 'Removed stale worker PID file.\n'
  exit 0
fi

if ! kill -0 "$pid" 2>/dev/null; then
  rm -f "$PID_FILE"
  printf 'Removed stale worker PID file.\n'
  exit 0
fi

command_line="$(ps -p "$pid" -o command= 2>/dev/null || true)"
if [[ "$command_line" != *"celery"* || "$command_line" != *"config.celery"* || "$command_line" != *"provisioning"* ]]; then
  printf 'Refusing to stop PID %s because it is not this project worker.\n' "$pid" >&2
  exit 1
fi

kill -TERM "$pid"
for _ in {1..20}; do
  if ! kill -0 "$pid" 2>/dev/null; then
    rm -f "$PID_FILE"
    printf 'Stopped background worker (PID %s).\n' "$pid"
    exit 0
  fi
  sleep 0.5
done

printf 'Worker did not stop gracefully; terminating PID %s.\n' "$pid" >&2
kill -KILL "$pid" 2>/dev/null || true
rm -f "$PID_FILE"
printf 'Stopped background worker (PID %s).\n' "$pid"
