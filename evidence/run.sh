#!/usr/bin/env bash
set -euo pipefail
cd /tmp/y852c3c
label=$1
dest=$PWD/bin-$label
log=$PWD/evidence/$label
tree=${CENSUS_TREE:-/workspace/scratch/y852c3c-adapted}
export CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_ASSERT_NO_OUTPUT=1 GOMAXPROCS=2
unset ADAMIC_MODULE_FORMAT ADAMIC_PROGRAM_REGION ADAMIC_OWNERSHIP_QUERY LATENT_SPECULATIVE LATENT_ONLY_FILE LATENT_ROOT_MANIFEST
case "$2" in
probe)
    mode=$3
    tag=$mode
    if [[ $# == 4 ]]; then
        export LATENT_ONLY_FILE="$tree/$4"
        tag="$mode-${4##*/}"
    fi
    LATENT_FULL="$mode" "$dest/record-census" "$tree/src/compiler" "$log/record-probe-$tag.jsonl" > "$log/record-probe-$tag.log" 2>&1
    ;;
directory)
    LATENT_TIMES="$log/directory-times.jsonl" LATENT_FULL=0 "$dest/latent-census" "$tree/src/compiler" "$log/directory.jsonl" > "$log/directory.log" 2>&1
    ;;
entry)
    LATENT_FULL=0 "$dest/entry-census" "$tree/src/tsc/tsc.ts" "$log/entry.jsonl" > "$log/entry.log" 2>&1
    ;;
scan)
    "$dest/scan-all" "$tree/src/tsc/tsc.ts" > "$log/scan.json" 2> "$log/scan.log"
    ;;
full)
    name=${3##*/}
    if [[ -e "$log/$name-full.jsonl" && ${CENSUS_RETRY:-0} != 1 ]]; then
        exit 0
    fi
    driver=$dest/latent-census
    root=$tree/src/compiler
    if [[ "$3" != src/compiler/* ]]; then
        driver=$dest/entry-census
        root=$tree/src/tsc/tsc.ts
    fi
    LATENT_FULL=1 LATENT_ONLY_FILE="$tree/$3" LATENT_TIMES="$log/$name-full-times.jsonl" "$driver" "$root" "$log/$name-full.jsonl" > "$log/$name-full.log" 2>&1
    ;;
esac
