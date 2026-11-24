#!/usr/bin/env bash
# Verdict case: synthetic-declaration, files/out/emit.d.ts (byte mismatch)
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" declaration-byte "$@"
