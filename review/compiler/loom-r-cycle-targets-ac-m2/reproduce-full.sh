#!/bin/bash
set -euo pipefail
repo=/workspace/adamic
out=$repo/review/compiler/loom-r-cycle-targets-ac-m2
for spec in parent:67733635b0252eb9d3bb3645c72262c7b18361e1 source:ca09c7495135717d4c79007ee6d4da560c255a26; do
 label=${spec%%:*}; revision=${spec#*:}; tree=/workspace/cycle-targets-repro-$label
 timeout 90 git -C "$repo" worktree add --detach "$tree" "$revision" >> "$out/reproduce-worktrees.log" 2>&1
 timeout 120 git -C "$tree" -c protocol.file.allow=always -c submodule.cohere.url="$repo/cohere" submodule update --init >> "$out/reproduce-worktrees.log" 2>&1
 set +e
 (cd "$tree" && timeout 630 go test ./internal/fuzz -count=1 -parallel=4 -json -timeout=600s) > "$out/$label-fuzz-full.log" 2>&1
 status=$?
 set -e
 printf '%s %s exit=%s\n' "$label" "$revision" "$status" >> "$out/reproduce-full-results.log"
 # Keep exact historical trees until their outputs have been reviewed.
done
