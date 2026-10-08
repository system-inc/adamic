#!/usr/bin/env bash
# Intended for morning/evening cron. Every invocation measures a fresh adapted tree.
set -euo pipefail
repository=$(cd "$(dirname "$0")/../.." && pwd)
cd "$repository"
for prerequisite in stage3/apply.sh stage3/census/tool/main.go stage3/census/latent/make_overlay.py; do
    [[ -f $prerequisite ]] || { echo "missing prerequisite: $prerequisite (merge stage3-base/tsc-census)" >&2; exit 1; }
done
compiler_mode=${STAGE3_METER_COMPILER:-single}
case $compiler_mode in
    single | per-ref) ;;
    *) echo "invalid STAGE3_METER_COMPILER: $compiler_mode (expected single or per-ref)" >&2; exit 2 ;;
esac
runs=${STAGE3_METER_RUNS:-$repository/stage3/meter/runs}
mkdir -p "$runs"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
run=$(mktemp -d "$runs/$stamp.XXXXXX")
scratch=$(mktemp -d "${TMPDIR:-/tmp}/stage3-meter.XXXXXX")
printf 'run: %s\n' "$run"
git fetch origin main:refs/remotes/origin/main area/stage3:refs/remotes/origin/area/stage3 > "$run/fetch.log" 2>&1
main_commit=$(git rev-parse origin/main)
area_commit=$(git rev-parse origin/area/stage3)
checkout_commit=$(git rev-parse HEAD)
python3 - "$run/compiler-mode.json" "$compiler_mode" "$checkout_commit" "$main_commit" "$area_commit" <<'PYTHON'
import json, sys
path, mode, checkout, main, area = sys.argv[1:]
with open(path, 'w') as output:
    json.dump({'mode': mode, 'commits': {'main': main if mode == 'per-ref' else checkout,
                                      'area': area if mode == 'per-ref' else checkout}}, output, indent=2)
    output.write('\n')
PYTHON
build_census() {
    local source=$1 destination=$2 logs=$3
    mkdir -p "$destination"
    (cd "$source" && go build -o "$destination/census" ./stage3/census/tool) > "$logs/build.log" 2>&1
    python3 "$source/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$logs/latent-overlay.log" 2>&1
    gofmt -w "$destination/latent-overlay/"*.go
    python3 "$repository/stage3/meter/entry_overlay.py" "$destination/latent-overlay" "$destination/entry-overlay" > "$logs/entry-overlay.log" 2>&1
    gofmt -w "$destination/entry-overlay/"*.go
    (cd "$source" && go build -buildvcs=false -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool) > "$logs/entry-build.log" 2>&1
    (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$logs/latent-build.log" 2>&1
}
if [[ $compiler_mode == single ]]; then
    build_census "$repository" "$scratch" "$run"
fi
for label in main area; do
    commit=$main_commit
    [[ $label == main ]] || commit=$area_commit
    mkdir -p "$scratch/$label-source" "$run/$label"
    # Each ref supplies its own apply script and adaptations. Its generated
    # patch-set stays in this snapshot instead of modifying the working tree.
    binaries=$scratch
    if [[ $compiler_mode == per-ref ]]; then
        # Detached worktrees retain each compiler's exact sources and submodule pins.
        git worktree add --detach "$scratch/$label-source" "$commit" > "$run/$label/checkout.log" 2>&1
        git -C "$scratch/$label-source" -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1 > "$run/$label/submodules.log" 2>&1
        binaries=$scratch/$label-binaries
        build_census "$scratch/$label-source" "$binaries" "$run/$label"
    else
        git archive "$commit" stage3 > "$scratch/$label.tar"
        tar -xf "$scratch/$label.tar" -C "$scratch/$label-source"
    fi
    bash "$scratch/$label-source/stage3/apply.sh" "$scratch/$label-adapted" > "$run/$label/apply.log" 2>&1
    entry=$scratch/$label-adapted/src/tsc/tsc.ts
    [[ -f $entry ]] || { echo "missing tsc entry: $entry" >&2; exit 1; }
    mkdir -p "$run/$label/tsc"
    "$binaries/census" "$entry" "$run/$label/tsc/census.jsonl" > "$run/$label/tsc/census.log" 2>&1
    LATENT_ASSERT_NO_OUTPUT=1 "$binaries/entry-census" "$entry" "$run/$label/tsc/latent.jsonl" > "$run/$label/tsc/latent.log" 2>&1
    "$binaries/census" "$scratch/$label-adapted/src/compiler" "$run/$label/census.jsonl" > "$run/$label/census.log" 2>&1
    LATENT_ASSERT_NO_OUTPUT=1 "$binaries/latent-census" "$scratch/$label-adapted/src/compiler" "$run/$label/latent.jsonl" > "$run/$label/latent.log" 2>&1
done
python3 stage3/meter/report.py "$scratch/main-adapted" "$scratch/area-adapted" "$run" "$stamp" "$main_commit" "$area_commit"
gzip "$run/main/tsc/census.jsonl" "$run/area/tsc/census.jsonl" "$run/main/tsc/latent.jsonl" "$run/area/tsc/latent.jsonl" "$run/main/census.jsonl" "$run/area/census.jsonl" "$run/main/latent.jsonl" "$run/area/latent.jsonl"
cat "$run/report.md"
