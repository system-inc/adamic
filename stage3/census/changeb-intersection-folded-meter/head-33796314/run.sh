#!/usr/bin/env bash
set -uo pipefail
source /workspace/adamic-tools/env.sh
base=/workspace/adamic/stage3/census/changeb-intersection-folded-meter
out="$base/head-33796314"
python3 "$base/measure.py" "$out/c-resource.json" timeout --kill-after=60s 5400 env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/c-times.jsonl" LATENT_PROGRESS_FILE="$out/c-progress.json" /workspace/scratch/census-c /workspace/scratch/tree/src/compiler "$out/c.jsonl" > "$out/c.log" 2>&1
code=$?
echo "$code" > "$out/c-exit.txt"
exit "$code"
