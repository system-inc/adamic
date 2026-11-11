#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
out=$PWD/stage3/census/map-object-keys-63979123
scratch=/workspace/scratch/ff5eddz
mkdir -p "$scratch"
git worktree add --detach "$scratch/head-source" 63979123 > "$out/head-checkout.log" 2>&1
rmdir "$scratch/head-source/cohere"
ln -s /workspace/adamic/cohere "$scratch/head-source/cohere"
for label in parent head; do
 src=/workspace/adamic
 if [ "$label" = head ]; then src=$scratch/head-source; fi
 bin=$scratch/$label
 mkdir -p "$bin"
 python3 "$src/stage3/census/latent/make_overlay.py" "$src" "$bin/latent-overlay" > "$out/$label-overlay.log" 2>&1
 gofmt -w "$bin/latent-overlay/"*.go
 python3 "$src/stage3/meter/entry_overlay.py" "$bin/latent-overlay" "$bin/entry-overlay" > "$out/$label-entry-overlay.log" 2>&1
 gofmt -w "$bin/entry-overlay/"*.go
 (cd "$src" && go build -buildvcs=false -overlay="$bin/latent-overlay/overlay.json" -o "$bin/latent-census" ./stage3/census/latent/tool) > "$out/$label-build.log" 2>&1
 (cd "$src" && go build -buildvcs=false -overlay="$bin/entry-overlay/overlay.json" -o "$bin/entry-census" ./stage3/census/latent/tool) > "$out/$label-entry-build.log" 2>&1
 git -C "$src" rev-parse HEAD > "$out/$label-compiler-sha.txt"
 sha256sum "$bin/latent-census" "$bin/entry-census" > "$out/$label-binaries.sha256"
done
