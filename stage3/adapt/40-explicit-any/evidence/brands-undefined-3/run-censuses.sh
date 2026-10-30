#!/usr/bin/env bash
set -euo pipefail
export PYTHONDONTWRITEBYTECODE=1
cd /workspace/adamic
while [[ ! -f /workspace/scratch/brands3/oracles-done.json ]]; do sleep 2; done
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/brands3/census/latent-census /workspace/scratch/brands2/control-tree/src/compiler /workspace/scratch/brands3/census/control.jsonl > /workspace/scratch/brands3/census/control.log 2>&1 &
p1=$!
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /workspace/scratch/brands3/census/latent-census /workspace/scratch/brands3/retyped-tree/src/compiler /workspace/scratch/brands3/census/retyped.jsonl > /workspace/scratch/brands3/census/retyped.log 2>&1 &
p2=$!
status=0
wait "$p1" || status=$?
wait "$p2" || status=$?
[[ $status == 0 ]]
python3 stage3/adapt/40-explicit-any/census-delta.py /workspace/scratch/brands3/census/control.jsonl /workspace/scratch/brands3/census/retyped.jsonl stage3/adapt/40-explicit-any/evidence/brands-undefined-3 > /workspace/scratch/brands3/census/delta.log 2>&1
python3 stage3/adapt/40-explicit-any/split-previous-133.py stage3/adapt/40-explicit-any/evidence/brands-undefined /workspace/scratch/brands3/census/control.jsonl /workspace/scratch/brands3/census/retyped.jsonl stage3/adapt/40-explicit-any/evidence/brands-undefined-3 > /workspace/scratch/brands3/census/split.log 2>&1
printf '{"status":"pass"}\n' > /workspace/scratch/brands3/censuses-done.json
