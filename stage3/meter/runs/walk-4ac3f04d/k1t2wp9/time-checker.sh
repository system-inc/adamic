#!/usr/bin/env bash
set -euo pipefail
cd /workspace/adamic
source /workspace/adamic-tools/env.sh
out=$PWD/stage3/meter/runs/walk-4ac3f04d/k1t2wp9/timing
mkdir -p "$out"
timebin=/usr/bin/time
if [[ ! -x $timebin ]]; then
    mkdir -p /workspace/adamic-tools/gnu-time
    curl -fsSL https://deb.debian.org/debian/pool/main/t/time/time_1.9-0.2_amd64.deb -o /workspace/adamic-tools/gnu-time/time_1.9-0.2_amd64.deb
    dpkg-deb -x /workspace/adamic-tools/gnu-time/time_1.9-0.2_amd64.deb /workspace/adamic-tools/gnu-time/extracted
    timebin=/workspace/adamic-tools/gnu-time/extracted/usr/bin/time
    sha256sum /workspace/adamic-tools/gnu-time/time_1.9-0.2_amd64.deb > "$out/tool-package.sha256"
fi
"$timebin" --version > "$out/tool.txt"
sha256sum "$timebin" >> "$out/tool.txt"
for index in 1 2 3; do
    for sha in 4ac3f04d767ec7d2d4e9b41ae9cdc9c4c9f4f288 8502a46e9289d6abb673f1359aa12d66be659b55; do
        case $sha in
            4ac3*) binary=/workspace/adamic-tools/k1t2wp9-before/binaries/latent-census; sourcefile=/workspace/adamic-tools/k1t2wp9-before/adapted/src/compiler/checker.ts ;;
            8502*) binary=/tmp/adamic-gate/stage3-meter.IuyyCt/main-binaries/latent-census; sourcefile=/tmp/adamic-gate/stage3-meter.IuyyCt/main-adapted/src/compiler/checker.ts ;;
        esac
        sha256sum "$binary" >> "$out/binaries.sha256"
        output=$out/$sha-checker-$index
        command=(env LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 "$timebin" -v -o "$output.time" "$binary" "$sourcefile" "$output.jsonl")
        { printf '%q ' "${command[@]}"; printf '> %q 2>&1\n' "$output.log"; } >> "$out/commands.log"
        "${command[@]}" > "$output.log" 2>&1
        gzip "$output.jsonl"
        printf '%s sample %s complete\n' "$sha" "$index"
    done
done
