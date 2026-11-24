#!/usr/bin/env bash
# Verdict case: 012_expandoFunctionExpressionsWithDynamicNames2, files/unexpected.js (extra under @noEmit)
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" no-emit-write "$@"
