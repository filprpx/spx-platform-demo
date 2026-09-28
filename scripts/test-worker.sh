#!/usr/bin/env bash
set -Eeuo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
grep -q '127.0.0.1:6379/0' "$ROOT_DIR/scripts/worker-start.sh"
grep -q -- '--queues=provisioning' "$ROOT_DIR/scripts/worker-start.sh"
grep -q '.venv/bin/celery' "$ROOT_DIR/scripts/worker-start.sh"
grep -q 'PLATFORM_API_URL=http://localhost:8000' "$ROOT_DIR/scripts/worker-start.sh"
grep -q 'worker-start' "$ROOT_DIR/Makefile"
grep -q 'worker-stop' "$ROOT_DIR/Makefile"
grep -q 'worker-status' "$ROOT_DIR/Makefile"
grep -q 'worker-logs' "$ROOT_DIR/Makefile"
grep -q 'worker-start.sh' "$ROOT_DIR/Makefile"
grep -q 'worker-stop.sh' "$ROOT_DIR/Makefile"
grep -q 'worker.pid' "$ROOT_DIR/scripts/worker-start.sh"
grep -q 'worker.pid' "$ROOT_DIR/scripts/worker-stop.sh"
grep -q '^iac-pipeline:' "$ROOT_DIR/Makefile"
grep -q 'execute-pipeline.py' "$ROOT_DIR/Makefile"
grep -q "python3 'host-side Celery worker runtime'" "$ROOT_DIR/scripts/check-dependencies.sh"
grep -q "openssl 'local secret and worker-token generation'" "$ROOT_DIR/scripts/check-dependencies.sh"
grep -q "python3 venv" "$ROOT_DIR/scripts/check-dependencies.sh"

printf 'worker target checks passed.\n'
