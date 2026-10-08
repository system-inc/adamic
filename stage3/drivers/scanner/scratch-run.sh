#!/usr/bin/env bash
# Retains scout byte comparisons and a scanner_native evidence block for the meter.
set -euo pipefail
scanner=$(cd "$(dirname "$0")" && pwd)
exec python3 "$scanner/scratch-run.py" "$@"
