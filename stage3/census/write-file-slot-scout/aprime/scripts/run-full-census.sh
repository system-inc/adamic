#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
export TMPDIR=/workspace/scratch/write-file-slot-aprime/tmp
export GOTMPDIR="$TMPDIR"
export GOCACHE=/workspace/scratch/write-file-slot-measure/gocache
export LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 CENSUS_MODULE_FORMAT=CommonJSBundle
export LATENT_ONLY_FILE=/workspace/scratch/write-file-slot-aprime/candidate/src/compiler/sys.ts
export LATENT_TIMES=/workspace/scratch/write-file-slot-aprime/full-sys.times.jsonl
export GOMEMLIMIT=10GiB
unset LATENT_SPECULATIVE LATENT_RUN_ONLY LATENT_ROOT_MANIFEST
/workspace/scratch/write-file-slot-measure/census /workspace/scratch/write-file-slot-aprime/candidate/src/compiler /workspace/scratch/write-file-slot-aprime/full-sys.jsonl
