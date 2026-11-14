#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/path-mint-618/env.sh
scratch=/workspace/scratch/path-mint-618
while [ ! -f "$scratch/lane/execution.json" ]; do sleep 2; done
tree=$scratch/lane/adapted-tree
(cd "$tree" && node node_modules/hereby/bin/hereby.js dts) > "$scratch/evidence/after-dts.log" 2>&1
python3 "$scratch/evidence/compare-dts.py"
python3 "$scratch/evidence/check-source.py" > "$scratch/evidence/source-comparison.log"
node "$scratch/work/stage3/adapt/99-correctness-timer-closures/witness.cjs" "$tree" > "$scratch/evidence/after-witness.json"
cmp "$scratch/evidence/base-witness.json" "$scratch/evidence/after-witness.json"
node "$scratch/evidence/replay.cjs" "$tree" > "$scratch/evidence/final-replay.json"
python3 "$scratch/evidence/provenance.py"
