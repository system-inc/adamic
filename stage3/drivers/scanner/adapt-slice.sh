#!/usr/bin/env bash
set -euo pipefail
repository=$(cd "$(dirname "$0")/../../.." && pwd)
node "$repository/stage3/slice/apply-adaptations.cjs" "${1:?usage: adapt-slice.sh SLICE}"
