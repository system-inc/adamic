#!/usr/bin/env bash
# Intended for morning/evening cron. Every invocation measures a fresh adapted tree.
set -euo pipefail
repository=$(cd "$(dirname "$0")/../.." && pwd)
cd "$repository"
for prerequisite in stage3/apply.sh stage3/census/tool/main.go; do
    [[ -f $prerequisite ]] || { echo "missing prerequisite: $prerequisite (merge stage3-base/tsc-census)" >&2; exit 1; }
done
runs=${STAGE3_METER_RUNS:-$repository/stage3/meter/runs}
mkdir -p "$runs"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
run=$(mktemp -d "$runs/$stamp.XXXXXX")
scratch=$(mktemp -d "${TMPDIR:-/tmp}/stage3-meter.XXXXXX")
tree=$scratch/adapted
printf 'run: %s\n' "$run"
bash stage3/apply.sh "$tree" > "$run/apply.log" 2>&1
go build -o "$scratch/census" ./stage3/census/tool > "$run/build.log" 2>&1
"$scratch/census" "$tree/src/compiler" "$run/census.jsonl" > "$run/census.log" 2>&1
python3 stage3/meter/report.py "$tree" "$run" "$stamp"
gzip "$run/census.jsonl"
cat "$run/report.md"
