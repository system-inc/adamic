#!/usr/bin/env bash
# Run from any directory after build-entry.sh completes.
set -uo pipefail
mkdir -p /workspace/scratch/results
TIMEFORMAT=$'real %R\nuser %U\nsys %S'
for label in main fix; do
 { time timeout --kill-after=60s 5400 env LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 "/workspace/scratch/$label-binaries/entry-census" /workspace/scratch/tree/src/tsc/tsc.ts "/workspace/scratch/results/$label.jsonl" > "/workspace/scratch/results/$label.log" 2>&1; } 2> "/workspace/scratch/results/$label.time"
 result=$?
 printf '%s\n' "$result" > "/workspace/scratch/results/$label.exit"
 [[ $result == 0 ]] || exit "$result"
done
