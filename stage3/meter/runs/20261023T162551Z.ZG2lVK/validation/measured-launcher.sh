#!/usr/bin/env bash
# Intended for morning/evening cron. Every invocation measures a fresh adapted tree.
set -euo pipefail
SECONDS=0
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
latent_mode=${STAGE3_METER_LATENT:-full}
case $latent_mode in
    full) latent_full=1 ;;
    first-error) latent_full=0 ;;
    *) echo "invalid STAGE3_METER_LATENT: $latent_mode (expected full or first-error)" >&2; exit 2 ;;
esac
# Resolve every source pin before creating any artifacts. The area label is a
# compatibility alias for main; the retired area branch is never fetched.
git fetch origin main:refs/remotes/origin/main >&2
main_ref=${STAGE3_METER_MAIN_REF:-origin/main}
main_commit=$(git rev-parse --verify "${main_ref}^{commit}") || { echo "cannot pin main: $main_ref" >&2; exit 1; }
area_ref=${STAGE3_METER_AREA_REF:-$main_commit}
area_commit=$(git rev-parse --verify "${area_ref}^{commit}") || { echo "cannot pin area alias: $area_ref" >&2; exit 1; }
[[ $main_commit == "$area_commit" ]] || { echo "main and area alias resolve to different shas" >&2; exit 1; }
git cat-file -e "$main_commit^{tree}" || exit 1
runs=${STAGE3_METER_RUNS:-$repository/stage3/meter/runs}
mkdir -p "$runs"
stamp=$(date -u +%Y%m%dT%H%M%SZ)
run=$(mktemp -d "$runs/$stamp.XXXXXX")
scratch=$(mktemp -d "${TMPDIR:-/tmp}/stage3-meter.XXXXXX")
printf 'run: %s\n' "$run"
# apply deliberately requires an explicit upstream input on a cold cache.
# Bootstrap the exact recorded tag when the scheduler has not supplied one.
if [[ -z ${ADAMIC_TYPESCRIPT_SOURCE:-} && -f $repository/stage3/source.json ]]; then
    upstream_url=$(git show "$main_commit:stage3/source.json" | python3 -c 'import json; print(json.load(__import__("sys").stdin)["repository"])')
    upstream_tag=$(git show "$main_commit:stage3/source.json" | python3 -c 'import json; print(json.load(__import__("sys").stdin)["tag"])')
    git clone --depth 1 --branch "$upstream_tag" "$upstream_url" "$scratch/upstream" > "$run/upstream.log" 2>&1
    export ADAMIC_TYPESCRIPT_SOURCE=$scratch/upstream
fi
checkout_commit=$(git rev-parse HEAD)
python3 - "$run/compiler-mode.json" "$compiler_mode" "$checkout_commit" "$main_commit" "$area_commit" "$latent_mode" <<'PYTHON'
import json, sys
path, mode, checkout, main, area, latent_mode = sys.argv[1:]
with open(path, 'w') as output:
    json.dump({'mode': mode, 'latent_mode': latent_mode, 'commits': {'main': main if mode == 'per-ref' else checkout,
                                      'area': area if mode == 'per-ref' else checkout}}, output, indent=2)
    output.write('\n')
PYTHON
build_census() {
    local source=$1 destination=$2 logs=$3
    mkdir -p "$destination"
    (cd "$source" && go build -o "$destination/census" ./stage3/census/tool) > "$logs/build.log" 2>&1
    python3 "$repository/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$logs/latent-overlay.log" 2>&1
    gofmt -w "$destination/latent-overlay/"*.go
    python3 "$repository/stage3/meter/entry_overlay.py" "$destination/latent-overlay" "$destination/entry-overlay" > "$logs/entry-overlay.log" 2>&1
    gofmt -w "$destination/entry-overlay/"*.go
    (cd "$source" && go build -buildvcs=false -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool) > "$logs/entry-build.log" 2>&1
    (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$logs/latent-build.log" 2>&1
}
if [[ $compiler_mode == single ]]; then
    build_census "$repository" "$scratch" "$run"
fi
measurement_pids=()
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
        if ! git -C "$scratch/$label-source" -c protocol.file.allow=always \
            -c "url.file://$repository/cohere.insteadOf=https://github.com/system-inc/cohere.git" \
            -c "url.file://$repository/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git" \
            submodule update --init --recursive --depth 1 > "$run/$label/submodules.log" 2>&1; then
            # A later main may pin dependencies absent from this worker's cache.
            # Retry their recorded origins; Git still checks out each exact gitlink.
            git -C "$scratch/$label-source" submodule sync --recursive >> "$run/$label/submodules.log" 2>&1
            git -C "$scratch/$label-source" -c protocol.file.allow=always submodule update --init --recursive --depth 1 >> "$run/$label/submodules.log" 2>&1
        fi
        binaries=$scratch/$label-binaries
        build_census "$scratch/$label-source" "$binaries" "$run/$label"
        git -C "$scratch/$label-source" submodule status --recursive > "$run/$label/submodule-pins.txt"
        git -C "$scratch/$label-source" status --porcelain --untracked-files=no > "$run/$label/compiler-status.txt"
        [[ ! -s $run/$label/compiler-status.txt ]] || { echo "dirty pinned compiler: $label" >&2; exit 1; }
        sha256sum "$binaries/census" "$binaries/latent-census" "$binaries/entry-census" > "$run/$label/binaries.sha256"
    else
        git archive "$commit" stage3 > "$scratch/$label.tar"
        tar -xf "$scratch/$label.tar" -C "$scratch/$label-source"
    fi
    # Cold daily checkouts need the apply tool's lockfile-pinned packages.
    if [[ -f $scratch/$label-source/stage3/api/package-lock.json ]]; then
        (cd "$scratch/$label-source/stage3/api" && npm ci --ignore-scripts --no-audit --no-fund) > "$run/$label/packages.log" 2>&1
    fi
    bash "$scratch/$label-source/stage3/apply.sh" "$scratch/$label-adapted" > "$run/$label/apply.log" 2>&1
    entry=$scratch/$label-adapted/src/tsc/tsc.ts
    [[ -f $entry ]] || { echo "missing tsc entry: $entry" >&2; exit 1; }
    mkdir -p "$run/$label/tsc"
    # Independent copies share CPU time; preparation stays ordered for apply caches.
    (
    "$binaries/census" "$entry" "$run/$label/tsc/census.jsonl" > "$run/$label/tsc/census.log" 2>&1
    LATENT_FULL=$latent_full LATENT_ASSERT_NO_OUTPUT=1 "$binaries/entry-census" "$entry" "$run/$label/tsc/latent.jsonl" > "$run/$label/tsc/latent.log" 2>&1
    "$binaries/census" "$scratch/$label-adapted/src/compiler" "$run/$label/census.jsonl" > "$run/$label/census.log" 2>&1
    LATENT_FULL=$latent_full LATENT_ASSERT_NO_OUTPUT=1 "$binaries/latent-census" "$scratch/$label-adapted/src/compiler" "$run/$label/latent.jsonl" > "$run/$label/latent.log" 2>&1
    ) &
    measurement_pids+=("$!")
done
measurement_status=0
for pid in "${measurement_pids[@]}"; do
    wait "$pid" || measurement_status=$?
done
[[ $measurement_status == 0 ]] || exit "$measurement_status"
python3 stage3/meter/report.py "$scratch/main-adapted" "$scratch/area-adapted" "$run" "$stamp" "$main_commit" "$area_commit"
gzip "$run/main/tsc/census.jsonl" "$run/area/tsc/census.jsonl" "$run/main/tsc/latent.jsonl" "$run/area/tsc/latent.jsonl" "$run/main/census.jsonl" "$run/area/census.jsonl" "$run/main/latent.jsonl" "$run/area/latent.jsonl"
python3 - "$run" "$SECONDS" <<'PYTHON'
import json, pathlib, sys
run = pathlib.Path(sys.argv[1])
seconds = int(sys.argv[2])
(run / 'timing.json').write_text(json.dumps({'elapsed_seconds': seconds}) + '\n')
with (run / 'README.md').open('a') as output:
    output.write(f'\nElapsed: {seconds // 60}m {seconds % 60}s.\n')
PYTHON
cat "$run/report.md"
