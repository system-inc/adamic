#!/bin/bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
root=/tmp/meter-stall-20261105/slowdown
cd /tmp/meter-stall-20261105/source
for item in good:f87d7884 bad:1f7df6a2; do
 label=${item%%:*}
 commit=${item#*:}
 mkdir -p "$root/plain/$label"
 git checkout --detach "$commit" > "$root/plain/$label/checkout.log" 2>&1
 go build -buildvcs=false -o "$root/plain/$label/census" ./stage3/census/tool > "$root/plain/$label/build.log" 2>&1
 date -u +%Y-%m-%dT%H:%M:%SZ > "$root/plain/$label/start.txt"
 set +e
 timeout --signal=QUIT --kill-after=20s 412.72938069s "$root/plain/$label/census" /tmp/meter-stall-20261105/adapted/src/compiler/tracing.ts "$root/plain/$label/records.jsonl" > "$root/plain/$label/run.log" 2>&1
 status=$?
 set -e
 printf '%s\n' "$status" > "$root/plain/$label/exit.txt"
 date -u +%Y-%m-%dT%H:%M:%SZ > "$root/plain/$label/end.txt"
 [[ $status == 0 ]] || exit "$status"
done
