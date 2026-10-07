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
printf 'run: %s\n' "$run"
git fetch origin main:refs/remotes/origin/main area/stage3:refs/remotes/origin/area/stage3 > "$run/fetch.log" 2>&1
main_commit=$(git rev-parse origin/main)
area_commit=$(git rev-parse origin/area/stage3)
go build -o "$scratch/census" ./stage3/census/tool > "$run/build.log" 2>&1
for label in main area; do
    commit=$main_commit
    [[ $label == main ]] || commit=$area_commit
    mkdir -p "$scratch/$label-source" "$run/$label"
    # Each ref supplies its own apply script and adaptations. Its generated
    # patch-set stays in this snapshot instead of modifying the working tree.
    git archive "$commit" stage3 > "$scratch/$label.tar"
    tar -xf "$scratch/$label.tar" -C "$scratch/$label-source"
    bash "$scratch/$label-source/stage3/apply.sh" "$scratch/$label-adapted" > "$run/$label/apply.log" 2>&1
    "$scratch/census" "$scratch/$label-adapted/src/compiler" "$run/$label/census.jsonl" > "$run/$label/census.log" 2>&1
done
python3 stage3/meter/report.py "$scratch/main-adapted" "$scratch/area-adapted" "$run" "$stamp" "$main_commit" "$area_commit"
gzip "$run/main/census.jsonl" "$run/area/census.jsonl"
cat "$run/report.md"
