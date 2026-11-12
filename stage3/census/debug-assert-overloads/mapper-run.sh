#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
while IFS='=' read -r name value; do
 case "$name" in LATENT_*|CENSUS_*) unset "$name";; esac
done < <(env)
scripts=/workspace/op2-report/stage3/census/debug-assert-overloads
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream GOMEMLIMIT=3GiB LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1
export LATENT_ONLY_FILE=/workspace/orm-adapted/src/compiler/transformers/destructuring.ts
set +e
python3 "$scripts/run.py" "$scripts/mapper-probe.log" "$scripts/mapper-probe.time.json" /workspace/dao-mapper/entry-census /workspace/orm-adapted/src/tsc/tsc.ts "$scripts/mapper-probe.jsonl"
rc=$?
echo "mapper census child_exit=$rc"
exit "$rc"
