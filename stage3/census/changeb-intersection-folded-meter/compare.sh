#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
cd /workspace/adamic
out=stage3/census/changeb-intersection-folded-meter
node "$out/direct-callers.cjs" > "$out/direct-callers.json"
python3 stage3/adapt/40-explicit-any/census-delta.py "$out/a.jsonl" "$out/b.jsonl" "$out" > "$out/delta.log" 2>&1
python3 "$out/analyze.py" > "$out/analysis.log" 2>&1
