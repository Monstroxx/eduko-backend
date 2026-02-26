#!/usr/bin/env bash
# Start the Eduko backend (development/manual mode).
# For production, use the systemd unit: scripts/eduko.service
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"

cd "$BACKEND_DIR"

# Defaults — override via env
export JWT_SECRET="${JWT_SECRET:-eduko-dev-secret-change-in-prod}"
export DATABASE_URL="${DATABASE_URL:-postgres://eduko:eduko@localhost:5432/eduko?sslmode=disable}"
export PORT="${PORT:-8080}"
export CORS_ORIGINS="${CORS_ORIGINS:-*}"

echo "[eduko] Starting backend on :${PORT}..."
exec ./eduko
