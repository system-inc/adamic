#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
out=$PWD/stage3/census/map-object-keys-63979123
adapted=/workspace/scratch/ff5eddz-adapted
for mode in tracing entry; do
 for side in parent head; do
  bin=/workspace/scratch/ff5eddz/$side
  export LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/$mode-$side-units.jsonl"
  if [ "$mode" = tracing ]; then
   export LATENT_FULL=1 LATENT_ONLY_FILE="$adapted/src/compiler/tracing.ts"
   python3 "$out/measure.py" "$out/$mode-$side" "$bin/latent-census" "$adapted/src/compiler" "$out/$mode-$side.jsonl"
  else
   export LATENT_FULL=0
   unset LATENT_ONLY_FILE
   python3 "$out/measure.py" "$out/$mode-$side" "$bin/entry-census" "$adapted/src/tsc/tsc.ts" "$out/$mode-$side.jsonl"
  fi
  bash "$out/checkpoint.sh" "$mode-$side"
 done
 base="$out/$mode-parent.jsonl"; tip="$out/$mode-head.jsonl"
 if [ -f "$base.gz" ]; then base="$base.gz"; fi
 if [ -f "$tip.gz" ]; then tip="$tip.gz"; fi
 python3 "$out/measure.py" "$out/$mode-pair" python3 stage3/census/pair/pair.py "$base" "$tip" "$adapted" --typescript stage3/api/node_modules/typescript --json "$out/$mode-delta.json" --report "$out/$mode-pair.md"
done
