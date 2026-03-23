#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PID_PATH="$ROOT/logs/server.pid"
STOPPED=0

if [[ -f "$PID_PATH" ]]; then
  PID="$(tr -d '[:space:]' < "$PID_PATH")"
  if [[ -n "$PID" ]] && kill -0 "$PID" >/dev/null 2>&1; then
    kill "$PID" >/dev/null 2>&1 || true
    sleep 1
    if kill -0 "$PID" >/dev/null 2>&1; then
      kill -9 "$PID" >/dev/null 2>&1 || true
    fi
    STOPPED=1
  fi
  rm -f "$PID_PATH"
fi

if [[ "$STOPPED" -eq 0 ]] && command -v lsof >/dev/null 2>&1; then
  PORT_PID="$(lsof -ti tcp:8080 -sTCP:LISTEN 2>/dev/null | head -n 1 || true)"
  if [[ -n "$PORT_PID" ]]; then
    kill "$PORT_PID" >/dev/null 2>&1 || true
    STOPPED=1
  fi
fi

if [[ "$STOPPED" -eq 0 ]] && command -v pkill >/dev/null 2>&1; then
  if pkill -f "/cmd/server" >/dev/null 2>&1; then
    STOPPED=1
  fi
fi

if [[ "$STOPPED" -eq 1 ]]; then
  echo "Online Judge stopped."
else
  echo "No running Online Judge process found."
fi
