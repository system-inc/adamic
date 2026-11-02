#!/usr/bin/env bash
set -uo pipefail
source /workspace/adamic-tools/env.sh
scratch=/workspace/scratch
out=/workspace/scratch/evidence/stage3/census/changeb-intersection-meter
cd /workspace/scratch/compiler-A
python3 stage3/meter/box.py "$out/box.json" start
cat /sys/fs/cgroup/memory.max > "$out/memory.max"
run() {
  local side=$1
  LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/$side-times.jsonl" LATENT_PROGRESS_FILE="$out/$side-progress.json" \
    python3 "$out/run-resource.py" "$out/$side-resource.json" timeout --kill-after=60s 5400 "$scratch/census-$side" "$scratch/tree/src/compiler" "$out/$side.jsonl" > "$out/$side.log" 2>&1
  local status=$?
  printf '%s\n' "$status" > "$out/$side-exit.txt"
  return "$status"
}
run A & p1=$!
run B & p2=$!
status=0
wait "$p1" || status=$?
wait "$p2" || status=$?
python3 stage3/meter/box.py "$out/box.json" end
exit "$status"
