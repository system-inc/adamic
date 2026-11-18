#!/usr/bin/env bash
set -u
label=$1
log=$2
shift 2
printf '%q ' "$@" > "$log.command"
printf '\n' >> "$log.command"
"$@" > "$log" 2>&1
rc=$?
if (( rc == 0 )); then
    printf 'PASS %s (child exit %d)\n' "$label" "$rc"
else
    printf 'FAIL %s (child exit %d)\n' "$label" "$rc"
fi
printf '%d\n' "$rc" > "$log.exit"
exit "$rc"
