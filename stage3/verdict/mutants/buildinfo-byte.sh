#!/usr/bin/env bash
# Verdict case: synthetic-incremental, files/out/tsconfig.tsbuildinfo (byte mismatch)
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" buildinfo-byte "$@"
