#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/path-mint-618/env.sh
scratch=/workspace/scratch/path-mint-618
node "$scratch/evidence/stock.cjs" "$scratch/base" "$scratch/evidence/base-stock" > "$scratch/evidence/base-stock.log" 2>&1
bash "$scratch/evidence/run-census.sh" "$scratch/base" base
python3 "$scratch/evidence/check-types.py" "$scratch/base" "$scratch/evidence/base-types.log"
(cd "$scratch/base" && node node_modules/hereby/bin/hereby.js dts) > "$scratch/evidence/base-dts.log" 2>&1
