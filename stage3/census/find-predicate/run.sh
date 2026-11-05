#!/usr/bin/env bash
set -euo pipefail
source /workspace/adamic-tools/env.sh
export GOPROXY='https://proxy.golang.org|direct'
R=/workspace/scratch
B=$R/find-base
(cd "$B/stage3/api" && npm ci --ignore-scripts --no-audit --no-fund) > "$R/find-packages.log" 2>&1
ADAMIC_TYPESCRIPT_SOURCE="$R/find-upstream" bash "$B/stage3/apply.sh" "$R/find-adapted" > "$R/find-apply.log" 2>&1
python3 "$B/stage3/meter/box.py" "$R/find-box.json" start
for label in base head; do
 source=$R/find-$label
 git -C "$source" -c submodule.cohere.url=https://github.com/system-inc/cohere.git submodule update --init --recursive --depth 1 > "$R/find-$label-submodules.log" 2>&1
 git -C "$source" submodule status --recursive > "$R/find-$label-pins.txt"
 destination=$R/find-$label-bin
 mkdir -p "$destination"
 python3 "$source/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$R/find-$label-overlay.log" 2>&1
 gofmt -w "$destination/latent-overlay/"*.go
 (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$R/find-$label-build.log" 2>&1
 git -C "$source" status --porcelain --untracked-files=no > "$R/find-$label-status.txt"
 cat /sys/fs/cgroup/memory.events > "$R/find-$label-memory-before.txt"
 LATENT_TIMES="$R/find-$label-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 python3 "$R/find-measure.py" "$R/find-$label-resource.json" "$destination/latent-census" "$R/find-adapted/src/compiler" "$R/find-$label-latent.jsonl" > "$R/find-$label-latent.log" 2>&1
 cat /sys/fs/cgroup/memory.events > "$R/find-$label-memory-after.txt"
done
python3 "$B/stage3/meter/box.py" "$R/find-box.json" end
