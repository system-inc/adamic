#!/usr/bin/env bash
# Verdict case: 001_varianceCantBeStrictWhileStructureIsnt, files/varianceCantBeStrictWhileStructureIsnt.js (missing)
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" missing-file "$@"
