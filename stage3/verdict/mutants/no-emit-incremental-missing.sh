#!/usr/bin/env bash
# Verdict case: synthetic-noemit-incremental, missing files/tsconfig.tsbuildinfo.
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" no-emit-incremental-missing "$@"
