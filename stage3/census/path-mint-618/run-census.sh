#!/usr/bin/env bash
set -euo pipefail
source /workspace/scratch/path-mint-618/env.sh
scratch=/workspace/scratch/path-mint-618
tree=$1
name=$2
if [ "$name" != base ]; then
  while [ ! -f "$scratch/lane/execution.json" ]; do sleep 2; done
fi
export GOMAXPROCS=1
LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_ONLY_FILE="$tree/src/compiler/sys.ts" CENSUS_MODULE_FORMAT=CommonJSBundle "$scratch/census" "$tree/src/compiler" "$scratch/evidence/$name.jsonl" > "$scratch/evidence/$name-census.log" 2>&1
