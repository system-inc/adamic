#!/usr/bin/env bash
set -euo pipefail
stage=$(cd "$(dirname "$0")" && pwd)
repo=$(cd "$stage/../../.." && pwd)
before=$(realpath "$1")
head=$(realpath "$2")
out=$(python3 -c 'from pathlib import Path; import sys; print(Path(sys.argv[1]).resolve())' "$3")
pin=0679623c2aa66cce5144ba0f894f020e3c5c6d99
if [[ $(git -C "$repo" rev-parse HEAD) != "$pin" && $(git -C "$repo" cat-file -p HEAD | sed -n '2s/^parent //p') != "$pin" ]]; then
    echo "Exposure replay requires compiler pin $pin" >&2
    exit 1
fi
mkdir -p "$out"
cd "$repo"
overlay=$(python3 stage3/census/latent/make_overlay.py "$repo" "$out/overlay")
go build -overlay "$overlay" -o "$out/replay" ./stage3/census/latent/replay/worker
export CENSUS_MODULE_FORMAT=CommonJSBundle LATENT_FULL=1 LATENT_ASSERT_NO_OUTPUT=1 LATENT_TAIL=1 GOMEMLIMIT=${GOMEMLIMIT:-4GiB}
workers_seconds=${FAMILY12_REPLAY_WORKERS_SECONDS:-600}
if [[ ! "$workers_seconds" =~ ^[1-9][0-9]*$ ]]; then
    echo "Invalid exact-unit replay worker deadline: $workers_seconds" >&2
    exit 2
fi
worker_deadline=$((SECONDS + workers_seconds))
for side in before head; do
    (
        tree=$before
        if [[ $side == head ]]; then tree=$head; fi
        if [[ -f "$tree/source.commit" ]]; then
            cp "$tree/source.commit" "$out/$side.source"
        else
            git -C "$tree" rev-parse HEAD > "$out/$side.source"
        fi
        node -e '
            const side = process.argv[2];
            require(process.argv[1]).forEach((row, index) => {
                const anchor = side === "before" ? row.beforeAnchor : row.headAnchor;
                console.log([index, ...anchor].join("\t"));
            });
        ' "$stage/exposure-selection.json" "$side" > "$out/$side.anchors"
        while IFS=$'\t' read -r index where kind reason; do
            remaining=$((worker_deadline - SECONDS))
            status=124
            if ((remaining > 0)); then
                status=0
                timeout --kill-after=2s "${remaining}s" "$out/replay" -project "$tree/src/compiler" -where "$tree/src/$where" -kind "$kind" -reason "$reason" \
                    > "$out/$side-$index.json" 2>"$out/$side-$index.log" || status=$?
            fi
            if ((status == 124 || status == 137)); then
                echo "Exact-unit replay worker phase exceeded ${workers_seconds}s ($side anchor $index)" >&2
                exit 124
            fi
            if ((status != 0)); then
                echo "Exact-unit replay worker phase failed ($side anchor $index, exit $status)" >&2
                exit "$status"
            fi
        done <"$out/$side.anchors"
    ) &
    if [[ "$side" == before ]]; then before_job=$!; else head_job=$!; fi
done
status=0
wait "$before_job" || status=$?
wait "$head_job" || { head_status=$?; if ((status == 0)); then status=$head_status; fi; }
if ((status != 0)); then exit "$status"; fi
node "$stage/check-exposures.cjs" "$out"
