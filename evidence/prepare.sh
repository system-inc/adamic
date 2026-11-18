#!/usr/bin/env bash
set -euo pipefail
top=/tmp/y852c3c
export PATH=/workspace/adamic-tools/go/bin:/workspace/adamic-tools/bin:$PATH
export GOCACHE=/workspace/scratch/y852c3c-go/cache GOMODCACHE=$top/gomodcache GOMAXPROCS=3 GOTMPDIR=/workspace/scratch/y852c3c-go/tmp
label=$1
source=$top/$label
dest=$top/bin-$label
log=$top/evidence/$label
cd "$source"
mkdir -p "$dest"
git rev-parse HEAD > "$log/commit.txt"
git submodule status --recursive > "$log/submodule-pins.txt"
python3 stage3/census/latent/make_overlay.py "$source" "$dest/latent-overlay" > "$log/latent-overlay.log" 2>&1
gofmt -w "$dest/latent-overlay/"*.go
python3 stage3/meter/entry_overlay.py "$dest/latent-overlay" "$dest/entry-overlay" > "$log/entry-overlay.log" 2>&1
gofmt -w "$dest/entry-overlay/"*.go
go build -buildvcs=false -overlay="$dest/latent-overlay/overlay.json" -o "$dest/latent-census" ./stage3/census/latent/tool > "$log/latent-build.log" 2>&1
go build -buildvcs=false -overlay="$dest/entry-overlay/overlay.json" -o "$dest/entry-census" ./stage3/census/latent/tool > "$log/entry-build.log" 2>&1
python3 stage3/census/scan-all/make_overlay.py "$dest/scan-overlay" > "$log/scan-overlay.log" 2>&1
go build -buildvcs=false -overlay="$dest/scan-overlay/overlay.json" -o "$dest/scan-all" ./stage3/census/scan-all/tool > "$log/scan-build.log" 2>&1
sha256sum "$dest/"{latent-census,entry-census,scan-all} > "$log/binaries.sha256"
