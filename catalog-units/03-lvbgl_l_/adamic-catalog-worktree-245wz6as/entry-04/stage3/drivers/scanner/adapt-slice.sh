#!/usr/bin/env bash
set -euo pipefail
repository=$(cd "$(dirname "$0")/../../.." && pwd)
api=${SLICE_TYPESCRIPT:?set SLICE_TYPESCRIPT to stock TypeScript 6.0.3 lib/typescript.js}
NODE_PATH="$(dirname "$(dirname "$(dirname "$api")")")" \
  node "$repository/stage3/adapt/42-scanner-any/adapt.cjs" "${1:?usage: adapt-slice.sh SLICE}"
node "$repository/stage3/slice/apply-adaptations.cjs" "${1:?usage: adapt-slice.sh SLICE}"
