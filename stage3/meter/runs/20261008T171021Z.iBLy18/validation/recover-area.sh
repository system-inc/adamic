#!/usr/bin/env bash
set -euo pipefail
repository=/workspace/stage3-meter-progress
run=$repository/stage3/meter/runs/20261008T171021Z.iBLy18
scratch=/workspace/stage3-meter-progress-tmp/stage3-meter.A1peqz
area_commit=45487a809f89885a3fc651cd590e7dabf31362dc
cd "$repository"
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

label=area
mkdir -p "$run/$label"
git worktree add --detach "$scratch/$label-source" "$area_commit" > "$run/$label/checkout.log" 2>&1
git -C "$scratch/$label-source" -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1 > "$run/$label/submodules.log" 2>&1
binaries=$scratch/$label-binaries
build_census "$scratch/$label-source" "$binaries" "$run/$label"
bash "$scratch/$label-source/stage3/apply.sh" "$scratch/$label-adapted" > "$run/$label/apply.log" 2>&1
entry=$scratch/$label-adapted/src/tsc/tsc.ts
[[ -f $entry ]]
mkdir -p "$run/$label/tsc"
"$binaries/census" "$entry" "$run/$label/tsc/census.jsonl" > "$run/$label/tsc/census.log" 2>&1
LATENT_ASSERT_NO_OUTPUT=1 "$binaries/entry-census" "$entry" "$run/$label/tsc/latent.jsonl" > "$run/$label/tsc/latent.log" 2>&1
"$binaries/census" "$scratch/$label-adapted/src/compiler" "$run/$label/census.jsonl" > "$run/$label/census.log" 2>&1
LATENT_ASSERT_NO_OUTPUT=1 "$binaries/latent-census" "$scratch/$label-adapted/src/compiler" "$run/$label/latent.jsonl" > "$run/$label/latent.log" 2>&1
