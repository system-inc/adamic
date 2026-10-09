#!/usr/bin/env bash
# Run on the Mac; --once is useful for supervised startup.
set -euo pipefail
here=$(cd "$(dirname "$0")" && pwd)
exec python3 "$here/idle-jobs.py" "$@"
