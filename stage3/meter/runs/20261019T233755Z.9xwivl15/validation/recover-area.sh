#!/usr/bin/env bash
# Resume after the retained pre-measurement overlay/input failures.
set -euo pipefail
repository=/workspace/adamic
cd "$repository"
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
export ADAMIC_TYPESCRIPT_SOURCE=/tmp/badrnrf-inputs/TypeScript
run=/workspace/adamic/stage3/meter/runs/20261019T233755Z.9xwivl15
scratch=/tmp/adamic-gate/stage3-meter.RXmuND
stamp=20261019T233755Z
main_commit=295df3bc6aa3d47933ccf675da1c33e745856313
area_commit=$main_commit
latent_full=0
build_census() {
    local source=$1 destination=$2 logs=$3
    mkdir -p "$destination"
    (cd "$source" && go build -o "$destination/census" ./stage3/census/tool) > "$logs/build.log" 2>&1
    python3 "/workspace/adamic/stage3/meter/runs/20261019T233755Z.9xwivl15/validation/make-overlay-run.py" "$source" "$destination/latent-overlay" > "$logs/latent-overlay.log" 2>&1
    gofmt -w "$destination/latent-overlay/"*.go
    python3 "$repository/stage3/meter/entry_overlay.py" "$destination/latent-overlay" "$destination/entry-overlay" > "$logs/entry-overlay.log" 2>&1
    gofmt -w "$destination/entry-overlay/"*.go
    (cd "$source" && go build -buildvcs=false -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool) > "$logs/entry-build.log" 2>&1
    (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$logs/latent-build.log" 2>&1
}
label=area
commit=$area_commit
mkdir -p "$scratch/$label-source" "$run/$label"
git worktree add --detach "$scratch/$label-source" "$commit" > "$run/$label/checkout.log" 2>&1
git -C "$scratch/$label-source" -c protocol.file.allow=always -c submodule.cohere.url=file:///workspace/adamic/cohere -c url.file:///workspace/adamic/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git submodule update --init --recursive --depth 1 > "$run/$label/submodules.log" 2>&1
ln -s /tmp/badrnrf-inputs/api/node_modules "$scratch/$label-source/stage3/api/node_modules"
binaries=$scratch/$label-binaries
build_census "$scratch/$label-source" "$binaries" "$run/$label"
bash "$scratch/$label-source/stage3/apply.sh" "$scratch/$label-adapted" > "$run/$label/apply.log" 2>&1
entry=$scratch/$label-adapted/src/tsc/tsc.ts
[[ -f $entry ]]
mkdir -p "$run/$label/tsc"
"$binaries/census" "$entry" "$run/$label/tsc/census.jsonl" > "$run/$label/tsc/census.log" 2>&1
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$binaries/entry-census" "$entry" "$run/$label/tsc/latent.jsonl" > "$run/$label/tsc/latent.log" 2>&1
"$binaries/census" "$scratch/$label-adapted/src/compiler" "$run/$label/census.jsonl" > "$run/$label/census.log" 2>&1
LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$binaries/latent-census" "$scratch/$label-adapted/src/compiler" "$run/$label/latent.jsonl" > "$run/$label/latent.log" 2>&1
