#!/usr/bin/env bash
set -uo pipefail
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/sourcefile-like-position-evidence/tmp
out=stage3/census/sourcefile-like-position
latent=/workspace/scratch/sourcefile-like-position-evidence/latent
for side in main trial; do
 if [ "$side" = main ]; then tree=/workspace/scratch/sourcefile-like-position-main-tree-2; else tree=/workspace/scratch/sourcefile-like-position-trial-tree; fi
 for name in scanner types; do
  LATENT_FULL=1 LATENT_ONLY_FILE="$tree/src/compiler/$name.ts" LATENT_TIMES="$PWD/$out/$side-$name-times.jsonl" "$latent" "$tree/src/compiler" "$PWD/$out/$side-$name.jsonl" > "$out/$side-$name.log" 2>&1
  status=$?
  echo "$status" > "$out/$side-$name.exit"
  if [ "$status" != 0 ]; then exit "$status"; fi
 done
done
