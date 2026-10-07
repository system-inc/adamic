#!/usr/bin/env bash
set -euo pipefail
repository=$(cd "$(dirname "$0")/../../.." && pwd)
slice_tree=$(realpath "${1:?usage: adapt-slice.sh RAW_SLICE}")
[[ -f "$slice_tree/slice.json" ]] || { echo 'expected a gathered declaration slice' >&2; exit 1; }
# Current compiler tips provide readiness and open numeric enums. Assertion
# call-site expansion remains needed until unknown parameter lowering lands.
for adaptation in 52 53 54 55 56 57 59 80 81 82 85; do
    scripts=("$repository"/stage3/adapt/"${adaptation}"-temporary-*/adapt.cjs)
    [[ ${#scripts[@]} -eq 1 && -f "${scripts[0]}" ]] || { echo "missing or ambiguous adaptation $adaptation" >&2; exit 1; }
    node "${scripts[0]}" "$slice_tree"
done
