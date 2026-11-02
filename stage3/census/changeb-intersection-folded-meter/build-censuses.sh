#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
for side in a b; do
  worktree=/workspace/scratch/compiler-$side
  overlay=/workspace/scratch/overlay-$side
  binary=/workspace/scratch/census-$side
  cd "$worktree"
  python3 stage3/census/latent/make_overlay.py "$worktree" "$overlay" > /workspace/scratch/overlay-$side.log 2>&1
  gofmt -w "$overlay"/*.go
  go build -buildvcs=false -overlay="$overlay"/overlay.json -o "$binary" ./stage3/census/latent/tool > /workspace/scratch/build-$side.log 2>&1
  sha256sum "$binary"
done
