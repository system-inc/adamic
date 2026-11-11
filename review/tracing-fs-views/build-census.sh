#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
scratch=/workspace/tracing-census-build
python3 -B stage3/census/latent/make_overlay.py /workspace/adamic "$scratch/full"
go build -overlay="$scratch/full/overlay.json" -o "$scratch/census" ./stage3/census/latent/tool
python3 -B stage3/meter/entry_overlay.py "$scratch/full" "$scratch/entry"
go build -overlay="$scratch/entry/overlay.json" -o "$scratch/entry-census" ./stage3/census/latent/tool
go build -o "$scratch/adamic" ./cmd/adamic
