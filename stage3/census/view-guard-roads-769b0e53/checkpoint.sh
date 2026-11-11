#!/usr/bin/env bash
set -euo pipefail
cd /workspace/roads-head
exec 9>/tmp/roads-checkpoint.lock
flock 9
folder=stage3/census/view-guard-roads-769b0e53
git add -f -- "$folder"
tree=$(git write-tree)
parent=769b0e534aa9106cec787a670d0243cd3734c9c0
if [[ -f /tmp/roads-checkpoint-sha ]]; then parent=$(cat /tmp/roads-checkpoint-sha); fi
sha=$(printf 'view guard roads census checkpoint: %s\n' "$1" | git commit-tree "$tree" -p "$parent")
set +e
git push origin "$sha:refs/heads/typescript/view-guard-roads-census-checkpoint" > "$folder/raw/checkpoint-$1.log" 2>&1
rc=$?
set -e
printf 'push_child_exit=%s sha=%s\n' "$rc" "$sha"
if [[ $rc -ne 0 ]]; then exit "$rc"; fi
printf '%s\n' "$sha" > /tmp/roads-checkpoint-sha
