#!/usr/bin/env bash
# Verdict case: 001_varianceCantBeStrictWhileStructureIsnt, files/varianceCantBeStrictWhileStructureIsnt.js (byte mismatch)
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" javascript-byte "$@"
