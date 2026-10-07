#!/usr/bin/env bash
set -euo pipefail
slice=$(cd "$(dirname "$0")" && pwd)
exec node "$slice/slice.cjs" "$@"
