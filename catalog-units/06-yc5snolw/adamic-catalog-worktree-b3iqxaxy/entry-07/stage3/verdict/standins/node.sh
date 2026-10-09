#!/usr/bin/env bash
set -euo pipefail
: "${STAGE3_VERDICT_NODE_TSC:?set to the adapted built/local/tsc.js}"
exec node "$STAGE3_VERDICT_NODE_TSC" "$@"
