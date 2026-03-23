#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ARCHIVE_PATH="${1:-$ROOT/artifacts/judge-images/algojudge-judge-images.tar}"

if [[ ! -f "$ARCHIVE_PATH" ]]; then
  echo "Archive not found: $ARCHIVE_PATH" >&2
  exit 1
fi

docker load -i "$ARCHIVE_PATH"
echo "Loaded judge images from:"
echo "  $ARCHIVE_PATH"
