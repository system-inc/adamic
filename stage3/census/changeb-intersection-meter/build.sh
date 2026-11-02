#!/usr/bin/env bash
set -euo pipefail
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
scratch=/workspace/scratch
cd "$scratch/compiler-A"
[[ $(git -C "$scratch/upstream" rev-parse HEAD) == 050880ce59e30b356b686bd3144efe24f875ebc8 ]]
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/upstream bash stage3/apply.sh /workspace/scratch/tree > "$scratch/apply.log" 2>&1
for side in A B; do
  worktree="$scratch/compiler-$side"
  # Both compiler pins use the same cohere gitlink as the setup clone.
  [[ $(git -C "$worktree" rev-parse HEAD:cohere) == $(git -C /workspace/adamic/cohere rev-parse HEAD) ]]
  rmdir "$worktree/cohere"
  ln -s /workspace/adamic/cohere "$worktree/cohere"
  cd "$worktree"
  python3 stage3/census/latent/make_overlay.py "$worktree" "$scratch/overlay-$side" > "$scratch/overlay-$side.log" 2>&1
  gofmt -w "$scratch/overlay-$side/"*.go
  go build -buildvcs=false -overlay="$scratch/overlay-$side/overlay.json" -o "$scratch/census-$side" ./stage3/census/latent/tool > "$scratch/build-$side.log" 2>&1
  sha256sum "$scratch/census-$side"
done
