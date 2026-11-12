#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
label=$1
shard=$2
scripts=/workspace/op2-report/stage3/census/overload-proof-3a7ea290
build=/workspace/op2-build-$label
# Clear unrelated census modes and mutants inherited from a shell.
while IFS='=' read -r name value; do
 case "$name" in LATENT_*|CENSUS_SHARD*) unset "$name";; esac
done < <(env)
export GOMEMLIMIT=3GiB
export LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1
export CENSUS_SHARDS=8 CENSUS_SHARD=$shard
export LATENT_TIMES="$build/shard-$shard-times.jsonl"
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/op2-upstream
set +e
python3 "$scripts/run.py" "$build/shard-$shard.log" "$build/shard-$shard.time.json" "$build/entry-census" /workspace/op2-adapted/src/tsc/tsc.ts "$build/shard-$shard.jsonl"
rc=$?
echo "measurement child_exit=$rc"
exit "$rc"
