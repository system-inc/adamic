#!/usr/bin/env bash
set -euo pipefail
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
for label in main fix; do
 source=/workspace/scratch/$label-source
 destination=/workspace/scratch/$label-binaries
 mkdir -p "$destination"
 git -C "$source" -c protocol.file.allow=always -c url.file:///workspace/adamic/cohere.insteadOf=https://github.com/system-inc/cohere.git -c url.file:///workspace/adamic/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git submodule update --init --recursive --depth 1 > "$destination/submodules.log" 2>&1
 python3 "$source/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$destination/latent-overlay.log" 2>&1
 gofmt -w "$destination/latent-overlay/"*.go
 python3 "$source/stage3/meter/entry_overlay.py" "$destination/latent-overlay" "$destination/entry-overlay" > "$destination/entry-overlay.log" 2>&1
 gofmt -w "$destination/entry-overlay/"*.go
 (cd "$source" && go build -buildvcs=false -overlay="$destination/entry-overlay/overlay.json" -o "$destination/entry-census" ./stage3/census/latent/tool) > "$destination/entry-build.log" 2>&1
 echo "$label entry build passed"
done
