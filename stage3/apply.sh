#!/usr/bin/env bash
set -euo pipefail
stage3=$(cd "$(dirname "$0")" && pwd)
exec python3 "$stage3/apply.py" "$@"
