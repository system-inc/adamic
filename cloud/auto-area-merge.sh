#!/usr/bin/env bash
# The watcher owns the durable queue and trusted integration checkout; Python also runs on macOS.
set -euo pipefail
here=$(cd "$(dirname "$0")/.." && pwd)
exec python3 "${here}/cloud/auto-area-merge.py" "$@"
