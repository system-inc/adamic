#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
out=stage3/census/map-object-keys-63979123
find "$out" -type f -size +1000000c ! -name '*.gz' -exec gzip {} \;
git add -f "$out"
tree=$(git write-tree)
sha=$(printf '%s\n' "Paired object-key Map census: $1" | git commit-tree "$tree" -p 73c047484f82f6f6546ff589214f3a0b7c8db9fb)
# Each checkpoint preserves the previous checkpoint as an extra parent so pushes are fast-forward.
if git show-ref --verify --quiet refs/heads/typescript/map-object-keys-census-checkpoint; then
 old=$(git rev-parse refs/heads/typescript/map-object-keys-census-checkpoint)
 sha=$(printf '%s\n' "Paired object-key Map census checkpoint: $1" | git commit-tree "$tree" -p "$old")
fi
git update-ref refs/heads/typescript/map-object-keys-census-checkpoint "$sha"
git push origin typescript/map-object-keys-census-checkpoint > "$out/checkpoint-$1-push.log" 2>&1
rc=$?
printf 'checkpoint push child exit=%s sha=%s\n' "$rc" "$sha"
exit "$rc"
