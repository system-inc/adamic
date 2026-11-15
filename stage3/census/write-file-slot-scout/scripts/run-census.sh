#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
scout_root=/workspace/scratch/write-file-slot-measure
export TMPDIR="$scout_root/tmp" GOTMPDIR="$scout_root/tmp" GOCACHE="$scout_root/gocache"
export LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle
export GOMEMLIMIT=10GiB
unset LATENT_SPECULATIVE LATENT_ROOT_MANIFEST LATENT_MUTANT_FIRST_ERROR_ONLY LATENT_MUTANT_NO_STUBS LATENT_RUN_ONLY
for road in main a-construction b-slot; do
 export LATENT_ONLY_FILE="$scout_root/$road/src/compiler/sys.ts"
 export LATENT_PROGRESS_FILE="$scout_root/$road.progress.json"
 export LATENT_TIMES="$scout_root/$road.times.jsonl"
 start=$SECONDS
 set +e
 "$scout_root/census" "$scout_root/$road/src/compiler" "$scout_root/$road.jsonl" > "$scout_root/$road.census.log" 2>&1
 code=$?
 set -e
 printf '%s\n' "$code" > "$scout_root/$road.census.exit"
 printf '%s %s %s\n' "$road" "$code" "$((SECONDS-start))"
 [ "$code" -eq 0 ] || exit "$code"
done
