#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
out=/workspace/scratch
python3 /workspace/scratch/compiler-A/stage3/meter/box.py "$out/box.json" start
set +e
python3 stage3/census/changeb-head1-meter/time-run.py "$out/A-resource.json" timeout --signal=TERM --kill-after=60s 90m env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/A-times.jsonl" LATENT_PROGRESS_FILE="$out/A-progress.json" "$out/census-A" "$out/tree/src/compiler" "$out/A.jsonl" > "$out/A.log" 2>&1 &
p1=$!
python3 stage3/census/changeb-head1-meter/time-run.py "$out/B-resource.json" timeout --signal=TERM --kill-after=60s 90m env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/B-times.jsonl" LATENT_PROGRESS_FILE="$out/B-progress.json" "$out/census-B" "$out/tree/src/compiler" "$out/B.jsonl" > "$out/B.log" 2>&1 &
p2=$!
wait "$p1"; a=$?
wait "$p2"; b=$?
printf 'A=%s\nB=%s\n' "$a" "$b" > "$out/census-exits.txt"
python3 /workspace/scratch/compiler-A/stage3/meter/box.py "$out/box.json" end
exit "$((a || b))"
