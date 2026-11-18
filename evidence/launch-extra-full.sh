#!/usr/bin/env bash
set -uo pipefail
cd /tmp/y852c3c
export GOMEMLIMIT=1350MiB GOGC=70
while IFS= read -r file; do
    bash run.sh base full "$file" &
    base=$!
    bash run.sh head full "$file" &
    head=$!
    wait "$base"
    base_rc=$?
    wait "$head"
    head_rc=$?
    printf '%s\t%s\t%s\n' "$file" "$base_rc" "$head_rc"
done < evidence/additional-target-files.txt
