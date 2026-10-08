#!/usr/bin/env bash
set -euo pipefail
lane=$(cd "$(dirname "$0")" && pwd)
exec python3 "$lane/run.py" "$@"
