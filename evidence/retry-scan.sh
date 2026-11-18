#!/usr/bin/env bash
set -euo pipefail
cd /tmp/y852c3c
until [[ $(wc -l < evidence/base/moduleNameResolver.ts-full.jsonl) -ge 3 ]] && [[ $(wc -l < evidence/head/moduleNameResolver.ts-full.jsonl) -ge 3 ]]; do sleep 10; done
export GOMEMLIMIT=1100MiB GOGC=70 GOMAXPROCS=2
for side in base head; do
    set +e
    "bin-$side/scan-all" /workspace/scratch/y852c3c-adapted/src/tsc/tsc.ts > "evidence/$side/scan.json" 2> "evidence/$side/scan.log"
    rc=$?
    set -e
    printf '%s\t%s\n' "$side" "$rc"
done
