#!/usr/bin/env bash
set -uo pipefail
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/sourcefile-like-position-evidence/tmp
mkdir -p "$TMPDIR"
evidence=stage3/census/sourcefile-like-position
python3 stage3/census/latent/make_overlay.py "$PWD" /workspace/scratch/sourcefile-like-position-evidence/overlay > "$evidence/overlay.log" 2>&1
status=$?
echo "$status" > "$evidence/overlay.exit"
if [ "$status" != 0 ]; then exit "$status"; fi
gofmt -w /workspace/scratch/sourcefile-like-position-evidence/overlay/*.go
go build -buildvcs=false -overlay=/workspace/scratch/sourcefile-like-position-evidence/overlay/overlay.json -o /workspace/scratch/sourcefile-like-position-evidence/latent ./stage3/census/latent/tool > "$evidence/build.log" 2>&1
status=$?
echo "$status" > "$evidence/build.exit"
exit "$status"
