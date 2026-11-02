#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
base=/workspace/adamic/stage3/census/changeb-intersection-folded-meter
out="$base/head-33796314"
inputs=/workspace/scratch/head-33796314-inputs
mkdir -p "$inputs"
gzip -dc "$base/a.jsonl.gz" > "$inputs/a.jsonl"
gzip -dc "$base/b.jsonl.gz" > "$inputs/b.jsonl"
python3 /workspace/scratch/compiler-a/stage3/adapt/40-explicit-any/census-delta.py "$inputs/a.jsonl" "$out/c.jsonl" "$out/vs-a" > "$out/delta-a.log" 2>&1
python3 /workspace/scratch/compiler-a/stage3/adapt/40-explicit-any/census-delta.py "$inputs/b.jsonl" "$out/c.jsonl" "$out/vs-5e5f57d3" > "$out/delta-b.log" 2>&1
python3 "$out/analyze.py" > "$out/analysis.log" 2>&1
