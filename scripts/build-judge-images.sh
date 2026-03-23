#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

docker build -t algojudge/judge-cpp:latest -f "$ROOT/docker/judger/cpp/Dockerfile" "$ROOT"
docker build -t algojudge/judge-java:latest -f "$ROOT/docker/judger/java/Dockerfile" "$ROOT"
docker build -t algojudge/judge-python:latest -f "$ROOT/docker/judger/python/Dockerfile" "$ROOT"
