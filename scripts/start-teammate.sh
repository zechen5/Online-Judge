#!/usr/bin/env bash
set -euo pipefail

NO_BROWSER=0
if [[ "${1:-}" == "--no-browser" ]]; then
  NO_BROWSER=1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOGS_DIR="$ROOT/logs"
STDOUT_PATH="$LOGS_DIR/server.log"
STDERR_PATH="$LOGS_DIR/server.err.log"
PID_PATH="$LOGS_DIR/server.pid"
ARCHIVE_PATH="$ROOT/artifacts/judge-images/algojudge-judge-images.tar"
HEALTHZ_URL="http://127.0.0.1:8080/healthz"
SITE_URL="http://127.0.0.1:8080/"
REQUIRED_IMAGES=(
  "algojudge/judge-cpp:latest"
  "algojudge/judge-java:latest"
  "algojudge/judge-python:latest"
)

assert_command() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "Missing required command: $1" >&2
    exit 1
  fi
}

open_url() {
  local url="$1"
  if command -v xdg-open >/dev/null 2>&1; then
    xdg-open "$url" >/dev/null 2>&1 || true
  elif command -v open >/dev/null 2>&1; then
    open "$url" >/dev/null 2>&1 || true
  fi
}

docker_quiet() {
  docker "$@" >/dev/null 2>&1
}

test_image_exists() {
  docker_quiet image inspect "$1"
}

ensure_docker_running() {
  if ! docker_quiet version; then
    echo "Docker is installed but not running. Please start Docker first." >&2
    exit 1
  fi
}

ensure_judge_images() {
  local missing=0
  for image in "${REQUIRED_IMAGES[@]}"; do
    if ! test_image_exists "$image"; then
      missing=1
      break
    fi
  done

  if [[ "$missing" -eq 0 ]]; then
    echo "Judge images already exist."
    return
  fi

  if [[ -f "$ARCHIVE_PATH" ]]; then
    echo "Judge image archive found. Loading local bundle..."
    bash "$ROOT/scripts/load-judge-images.sh" "$ARCHIVE_PATH"
  else
    echo "Judge image archive not found. Building images from Dockerfiles..."
    bash "$ROOT/scripts/build-judge-images.sh"
  fi

  for image in "${REQUIRED_IMAGES[@]}"; do
    if ! test_image_exists "$image"; then
      echo "Judge images are still missing: $image" >&2
      exit 1
    fi
  done
}

test_healthz() {
  curl -fsS "$HEALTHZ_URL" >/dev/null 2>&1
}

assert_port_available() {
  if command -v ss >/dev/null 2>&1; then
    if ss -ltn '( sport = :8080 )' 2>/dev/null | tail -n +2 | grep -q .; then
      if ! test_healthz; then
        echo "Port 8080 is already occupied by another process." >&2
        exit 1
      fi
    fi
    return
  fi

  if command -v lsof >/dev/null 2>&1; then
    if lsof -ti tcp:8080 -sTCP:LISTEN >/dev/null 2>&1; then
      if ! test_healthz; then
        echo "Port 8080 is already occupied by another process." >&2
        exit 1
      fi
    fi
  fi
}

start_server() {
  mkdir -p "$LOGS_DIR"
  rm -f "$STDOUT_PATH" "$STDERR_PATH" "$PID_PATH"
  (
    cd "$ROOT"
    OJ_JUDGE_EXECUTOR=docker nohup go run ./cmd/server >"$STDOUT_PATH" 2>"$STDERR_PATH" &
    echo $! >"$PID_PATH"
  )
}

assert_command go
assert_command docker
assert_command curl
ensure_docker_running
ensure_judge_images

if test_healthz; then
  echo "Server already running at $SITE_URL"
  if [[ "$NO_BROWSER" -eq 0 ]]; then
    open_url "$SITE_URL"
  fi
  exit 0
fi

assert_port_available
start_server

READY=0
for _ in $(seq 1 40); do
  sleep 0.5
  if test_healthz; then
    READY=1
    break
  fi
  if [[ -f "$PID_PATH" ]]; then
    PID="$(cat "$PID_PATH")"
    if ! kill -0 "$PID" >/dev/null 2>&1; then
      break
    fi
  fi
done

if [[ "$READY" -ne 1 ]]; then
  rm -f "$PID_PATH"
  echo "Server did not become ready in time. Check logs/server.log and logs/server.err.log" >&2
  exit 1
fi

echo "Online Judge is ready:"
echo "  $SITE_URL"
echo "Default admin:"
echo "  username: admin"
echo "  password: admin123456"
echo "Logs:"
echo "  $STDOUT_PATH"
echo "  $STDERR_PATH"
echo "PID:"
echo "  $PID_PATH"

if [[ "$NO_BROWSER" -eq 0 ]]; then
  open_url "$SITE_URL"
fi
