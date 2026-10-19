#!/bin/bash
set -euo pipefail
out=/workspace/adamic/review/compiler/loom-r-cycle-targets-ac-m2
for label in parent source; do
 tree=/workspace/cycle-targets-repro-$label
 git -C "$tree" rev-parse HEAD >> "$out/historical-revisions.log"
 git -C "$tree/cohere" ls-tree HEAD TypeScript >> "$out/historical-revisions.log"
 rmdir "$tree/cohere/TypeScript"
 ln -s /workspace/adamic/cohere/TypeScript "$tree/cohere/TypeScript"
 mv "$out/$label-fuzz-full.log" "$out/$label-fuzz-missing-submodule.log"
 set +e
 (cd "$tree" && timeout 630 go test ./internal/fuzz -count=1 -parallel=4 -json -timeout=600s) > "$out/$label-fuzz-full.log" 2>&1
 status=$?
 set -e
 printf '%s full fuzz exit=%s\n' "$label" "$status" >> "$out/reproduce-full-results-final.log"
done
