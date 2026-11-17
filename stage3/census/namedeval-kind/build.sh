#!/usr/bin/env bash
set -euo pipefail
case_name=$1
root=/workspace/scratch/namedeval-kind-$case_name
out=/workspace/scratch/namedeval-kind-runtime/$case_name
source /workspace/scratch/namedeval-kind-runtime/env.sh
export NODE_PATH="$root/stage3/api/node_modules"
mkdir -p "$out"
cd "$root"
python3 /workspace/scratch/namedeval-kind-runtime/run.py "$case_name-latent-overlay" python3 stage3/census/latent/make_overlay.py "$root" "$out/latent-overlay"
python3 /workspace/scratch/namedeval-kind-runtime/run.py "$case_name-latent-build" go build -trimpath -ldflags=-buildid= -buildvcs=false -overlay "$out/latent-overlay/overlay.json" -o "$out/latent" ./stage3/census/latent/tool/main.go
python3 /workspace/scratch/namedeval-kind-runtime/run.py "$case_name-scan-overlay" python3 stage3/census/scan-all/make_overlay.py "$out/scan-overlay"
python3 /workspace/scratch/namedeval-kind-runtime/run.py "$case_name-scan-build" go build -trimpath -ldflags=-buildid= -buildvcs=false -overlay "$out/scan-overlay/overlay.json" -o "$out/scan-all" ./stage3/census/scan-all/tool/main.go
