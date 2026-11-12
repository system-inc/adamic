#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
while IFS='=' read -r name value; do
 case "$name" in LATENT_*|CENSUS_*) unset "$name";; esac
done < <(env)
label=$1
scripts=/workspace/op2-report/stage3/census/create-builder-program
tree=/workspace/cbp-$label
if test "$label" = base; then tree=/workspace/orm-adapted; fi
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream GOMEMLIMIT=3GiB LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_SKIP_SCAN=1
export LATENT_ONLY_FILE="$tree/src/compiler/builder.ts
$tree/src/compiler/builderPublic.ts"
export LATENT_TIMES="$scripts/$label-unit-times.jsonl"
set +e
python3 "$scripts/run.py" "$scripts/$label-census.log" "$scripts/$label-census.time.json" /workspace/cbp-tool/entry-census "$tree/src/tsc/tsc.ts" "$scripts/$label-census.jsonl"
rc=$?
echo "$label census child_exit=$rc"
exit "$rc"
