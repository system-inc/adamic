#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
out=/workspace/adamic/stage3/census/changeb-intersection-folded-meter/head-33796314
cd /workspace/scratch/compiler-c
python3 stage3/census/latent/make_overlay.py /workspace/scratch/compiler-c /workspace/scratch/overlay-c > "$out/overlay.log" 2>&1
gofmt -w /workspace/scratch/overlay-c/*.go
go build -buildvcs=false -overlay=/workspace/scratch/overlay-c/overlay.json -o /workspace/scratch/census-c ./stage3/census/latent/tool > "$out/build.log" 2>&1
sha256sum /workspace/scratch/census-c > "$out/binary.sha256"
