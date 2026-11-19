#!/usr/bin/env bash
set -euo pipefail
d=/workspace/scratch/family5-measure
e=$d/report/stage3/typescript/family5-pragma-scout/evidence/measurement
for label in before after; do
    roots=$d/$label/src/compiler
    { time env LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_ONLY_FILE="$roots/parser.ts" "$d/latent-census" "$roots" "$d/$label-latent.jsonl" > "$e/$label-latent.log" 2>&1; } 2> "$e/$label-latent-time.txt"
    gzip -cn "$d/$label-latent.jsonl" > "$e/$label-latent.jsonl.gz"
done
