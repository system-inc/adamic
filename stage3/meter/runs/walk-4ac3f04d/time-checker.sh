#!/usr/bin/env bash
# Run only after both meter invocations finish successfully.
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
outdir=/workspace/adamic/stage3/meter/runs/walk-4ac3f04d/timing
mkdir -p "$outdir"
for sha in 76ea8b17ea0ed467ecb86f355c06f738fd262983 4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288; do
    case $sha in
        76ea*) scratch=/tmp/adamic-gate/stage3-meter.nfa1mc ;;
        4ac3*) scratch=/tmp/adamic-gate/stage3-meter.4OGO7t ;;
    esac
    binary=$scratch/main-binaries/latent-census
    sourcefile=$scratch/main-adapted/src/compiler/checker.ts
    sha256sum "$binary" >> "$outdir/binaries.sha256"
    for index in 1 2 3; do
        output=$outdir/$sha-checker-$index
        command=(env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 /workspace/adamic-tools/gnu-time/extracted/usr/bin/time -v -o "$output.time" "$binary" "$sourcefile" "$output.jsonl")
        { printf '%q ' "${command[@]}"; printf '> %q 2>&1\n' "$output.log"; } >> "$outdir/commands.log"
        "${command[@]}" > "$output.log" 2>&1
        gzip "$output.jsonl"
        printf '%s checker %s complete\n' "$sha" "$index"
    done
done
