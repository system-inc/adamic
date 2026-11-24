#!/usr/bin/env bash
# Verdict case: 001_varianceCantBeStrictWhileStructureIsnt, files/unexpected.js (extra)
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" extra-file "$@"
