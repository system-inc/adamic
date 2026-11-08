#!/usr/bin/env bash
set -uo pipefail
source /workspace/adamic-tools/env.sh
export ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/8dbaea5/typescript
measurement_root=/workspace/scratch/8dbaea5
run_step() {
    local step=$1
    shift
    "$@" > "$measurement_root/$step.log" 2>&1
    local code=$?
    printf '%s\n' "$code" > "$measurement_root/$step.exit"
    printf '%s_exit=%s\n' "$step" "$code"
    if [ "$code" != 0 ]; then exit "$code"; fi
}
cd /workspace/adamic
[ "$(git -C "$ADAMIC_TYPESCRIPT_SOURCE" rev-parse HEAD)" = 050880ce59e30b356b686bd3144efe24f875ebc8 ] || exit 2
run_step main-npm npm ci --prefix "$measurement_root/main-source/stage3/api"
run_step apply bash "$measurement_root/main-source/stage3/apply.sh" "$measurement_root/adapted"
mkdir -p "$measurement_root/checker-full"
export LATENT_FULL=1
export LATENT_ONLY_FILE="$measurement_root/adapted/src/compiler/checker.ts"
export LATENT_ASSERT_NO_OUTPUT=1
export LATENT_RELATION_PER_UNIT=0
export LATENT_SPECULATIVE=0
export LATENT_TIMES="$measurement_root/checker-full/unit-times.jsonl"
export LATENT_RELATION_TIMES="$measurement_root/checker-full/relation-times.jsonl"
export GOMEMLIMIT=8GiB
unset LATENT_ROOT_MANIFEST LATENT_MUTANT_FIRST_ERROR_ONLY LATENT_MUTANT_NO_ROLLBACK LATENT_MUTANT_FUNCTION LATENT_MUTANT_WHERE
python3 stage3/census/speculative/timed_run.py 6000 12288 "$measurement_root/checker-full/metrics.json" "$measurement_root/checker-full/run.log" "$measurement_root/entry-census" "$measurement_root/adapted/src/tsc/tsc.ts" "$measurement_root/checker-full/latent.jsonl" > "$measurement_root/checker-full/wrapper.log" 2>&1
code=$?
printf '%s\n' "$code" > "$measurement_root/checker-full/wrapper.exit"
printf 'measurement_exit=%s\n' "$code"
exit "$code"
