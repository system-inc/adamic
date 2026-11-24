#!/usr/bin/env bash
# Verdict case: synthetic-noemit-incremental, extra files/unexpected.js.
set -euo pipefail
exec python3 "$(dirname "${BASH_SOURCE[0]}")/standin.py" no-emit-incremental-js "$@"
