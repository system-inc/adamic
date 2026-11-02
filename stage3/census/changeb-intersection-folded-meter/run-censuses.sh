#!/usr/bin/env bash
set -uo pipefail
source /workspace/adamic-tools/env.sh
out=/workspace/adamic/stage3/census/changeb-intersection-folded-meter
for side in a b; do
  (
    python3 "$out/measure.py" "$out/$side-resource.json" timeout --kill-after=60s 5400 env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/$side-times.jsonl" LATENT_PROGRESS_FILE="$out/$side-progress.json" /workspace/scratch/census-$side /workspace/scratch/tree/src/compiler "$out/$side.jsonl" > "$out/$side.log" 2>&1
    echo "$?" > "$out/$side-exit.txt"
  ) &
  if [[ $side == a ]]; then pa=$!; else pb=$!; fi
done
wait "$pa"
wait "$pb"
