#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/overload-upstream
label=$1
out=/workspace/overload-build-$label
export LATENT_ASSERT_NO_OUTPUT=1 LATENT_FULL=0 LATENT_TIMES="$out/latent-times.jsonl"
unset LATENT_SPECULATIVE LATENT_MUTANT_FUNCTION LATENT_ONLY_FILE LATENT_ONLY_UNIT
set +e
python3 /workspace/overload-run.py "$out/latent.log" "$out/latent.time.json" "$out/entry-census" /workspace/overload-adapted/src/tsc/tsc.ts "$out/latent.jsonl"
rc=$?
echo "entry_census_child_exit=$rc"
exit "$rc"
