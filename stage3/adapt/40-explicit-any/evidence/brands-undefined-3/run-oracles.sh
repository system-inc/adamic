#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
export PYTHONDONTWRITEBYTECODE=1
cd /workspace/adamic
while [[ ! -f /workspace/scratch/brands3/preparation-done.json ]]; do sleep 2; done
cd /workspace/scratch/brands3/control-source
STAGE3_ADAPTED_TREE=/workspace/scratch/brands2/control-tree PYTHONPATH="$PWD/stage3/lane" python3 /workspace/scratch/brands2/measure-shards-release-views.py /workspace/scratch/brands3/control-shards --concurrency 4 > /workspace/scratch/brands3/control-shards.log 2>&1
cd /workspace/adamic
python3 /workspace/scratch/brands3/summarize-shards.py /workspace/scratch/brands3/control-source/stage3/lane /workspace/scratch/brands3/control-shards stage3/adapt/40-explicit-any/evidence/brands-undefined-3/control > /workspace/scratch/brands3/control-summary.log 2>&1
python3 /workspace/scratch/brands3/rebind-plan.py "$PWD/stage3/lane" > /workspace/scratch/brands3/retyped-manifest.json
STAGE3_ADAPTED_TREE=/workspace/scratch/brands3/retyped-tree PYTHONPATH="$PWD/stage3/lane" python3 /workspace/scratch/brands2/measure-shards-release-views.py /workspace/scratch/brands3/retyped-shards --concurrency 4 > /workspace/scratch/brands3/retyped-shards.log 2>&1
python3 /workspace/scratch/brands3/summarize-shards.py "$PWD/stage3/lane" /workspace/scratch/brands3/retyped-shards stage3/adapt/40-explicit-any/evidence/brands-undefined-3/retyped > /workspace/scratch/brands3/retyped-summary.log 2>&1
python3 stage3/adapt/40-explicit-any/sanction-undo-mutant.py prove /workspace/scratch/brands3/control-shards/stage3-lane-byte-comparisons stage3/adapt/40-explicit-any/evidence/brands-undefined-3 > /workspace/scratch/brands3/undo-mutant.log 2>&1
printf '{"status":"pass"}\n' > /workspace/scratch/brands3/oracles-done.json
