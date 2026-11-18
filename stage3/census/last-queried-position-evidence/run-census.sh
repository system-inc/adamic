#!/usr/bin/env bash
set -eu
tree=$1
label=$2
binary=$3
logs=$4
bash "$logs/run-child.sh" "$label full parser census" "$logs/$label-census.log" \
    env -u LATENT_SPECULATIVE LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle \
    LATENT_ONLY_FILE="$tree/src/compiler/parser.ts" LATENT_TIMES="$logs/$label-census-times.jsonl" \
    "$binary" "$tree/src/compiler" "$logs/$label-census.jsonl"
