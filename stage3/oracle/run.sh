#!/usr/bin/env bash
set -euo pipefail
oracle=$(cd "$(dirname "$0")" && pwd)
exec python3 "$oracle/run.py" "$@"
