#!/usr/bin/env bash
set -uo pipefail
out=/workspace/scratch/evidence/stage3/census/changeb-meter
python3 /workspace/scratch/change-a/stage3/meter/box.py "$out/box.json" start
sha256sum /workspace/scratch/census-a /workspace/scratch/census-b > "$out/binaries.sha256"
find /workspace/scratch/tree/src/compiler -type f \( -name '*.ts' -o -name '*.a' \) -print0 | sort -z | xargs -0 sha256sum > "$out/tree.sha256"
date -u +%FT%TZ > "$out/start.utc"
/usr/bin/time -v -o "$out/A.resources.txt" timeout --kill-after=60s 5400 env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/A.times.jsonl" /workspace/scratch/census-a /workspace/scratch/tree/src/compiler "$out/A.jsonl" > "$out/A.log" 2>&1 &
a=$!
/usr/bin/time -v -o "$out/B.resources.txt" timeout --kill-after=60s 5400 env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TIMES="$out/B.times.jsonl" /workspace/scratch/census-b /workspace/scratch/tree/src/compiler "$out/B.jsonl" > "$out/B.log" 2>&1 &
b=$!
wait "$a"; sa=$?
wait "$b"; sb=$?
printf '{"A":%d,"B":%d}\n' "$sa" "$sb" > "$out/exit.json"
date -u +%FT%TZ > "$out/end.utc"
python3 /workspace/scratch/change-a/stage3/meter/box.py "$out/box.json" end
