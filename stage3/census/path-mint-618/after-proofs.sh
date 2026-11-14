#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/path-mint-618/env.sh
scratch=/workspace/scratch/path-mint-618
while [ ! -f "$scratch/lane/adapted-tree/patch-set.md" ]; do sleep 2; done
tree=$scratch/lane/adapted-tree
node "$scratch/evidence/stock.cjs" "$tree" "$scratch/evidence/after-stock" > "$scratch/evidence/after-stock.log" 2>&1
node "$scratch/evidence/emit-all.cjs" "$tree" "$scratch/evidence/after-all-js.json" > "$scratch/evidence/after-all-js.log" 2>&1
bash "$scratch/evidence/run-census.sh" "$tree" after
python3 "$scratch/evidence/check-types.py" "$tree" "$scratch/evidence/after-types.log"
python3 "$scratch/evidence/mutant.py"
bash "$scratch/evidence/run-census.sh" "$scratch/mutant" mutant
python3 "$scratch/evidence/delta.py" > "$scratch/evidence/census-delta.log"
python3 "$scratch/evidence/check-js.py" > "$scratch/evidence/javascript-comparison.log"
