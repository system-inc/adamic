#!/usr/bin/env bash
set -euo pipefail
scanner=$(cd "$(dirname "$0")" && pwd)
exec python3 "$scanner/run.py" "$@"
