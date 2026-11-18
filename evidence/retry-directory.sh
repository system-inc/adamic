#!/usr/bin/env bash
set -euo pipefail
cd /tmp/y852c3c
until [[ $(wc -l < evidence/base/utilities.ts-full.jsonl) -ge 3 ]] &&
      [[ $(wc -l < evidence/head/utilities.ts-full.jsonl) -ge 3 ]] &&
      [[ $(wc -l < evidence/head/record-probe-1-utilities.ts.jsonl) -ge 3 ]]; do
    sleep 10
done
export GOMEMLIMIT=1250MiB GOGC=65
for side in base head; do
    set +e
    bash run.sh "$side" directory
    rc=$?
    set -e
    printf '%s\t%s\n' "$side" "$rc"
done
