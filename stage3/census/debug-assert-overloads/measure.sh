#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
while IFS='=' read -r name value; do
 case "$name" in LATENT_*|CENSUS_*) unset "$name";; esac
done < <(env)
scripts=/workspace/op2-report/stage3/census/debug-assert-overloads
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream
export GOMEMLIMIT=3GiB LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1
export LATENT_ONLY_FILE=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts
export LATENT_TIMES="$scripts/unit-times.jsonl"
set +e
python3 "$scripts/run.py" "$scripts/census.log" "$scripts/census.time.json" /workspace/orm-build-main/entry-census /workspace/orm-adapted/src/tsc/tsc.ts "$scripts/census.jsonl"
rc=$?
echo "full census child_exit=$rc"
exit "$rc"
