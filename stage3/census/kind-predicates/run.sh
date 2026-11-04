#!/usr/bin/env bash
set -euo pipefail
export GOPROXY='https://proxy.golang.org|direct'
source /workspace/adamic-tools/env.sh
root=/workspace/scratch
for label in base head; do
 source=$root/kp-$label
 if [[ ! -x $root/kp-$label-bin/latent-census ]]; then
 git -C "$source" -c protocol.file.allow=always -c url.file:///workspace/adamic/cohere.insteadOf=https://github.com/system-inc/cohere.git -c url.file:///workspace/adamic/cohere/TypeScript.insteadOf=https://github.com/system-inc/TypeScript.git submodule update --init --recursive --depth 1 > "$root/kp-$label-submodules.log" 2>&1
 git -C "$source" submodule status --recursive > "$root/kp-$label-pins.txt"
 destination=$root/kp-$label-bin
 mkdir -p "$destination"
 python3 "$source/stage3/census/latent/make_overlay.py" "$source" "$destination/latent-overlay" > "$root/kp-$label-overlay.log" 2>&1
 gofmt -w "$destination/latent-overlay/"*.go
 (cd "$source" && go build -buildvcs=false -overlay="$destination/latent-overlay/overlay.json" -o "$destination/latent-census" ./stage3/census/latent/tool) > "$root/kp-$label-build.log" 2>&1
 fi
 destination=$root/kp-$label-bin
 # Each build and run finishes before starting the next compiler.
 cat /sys/fs/cgroup/memory.events > "$root/kp-$label-memory-before.txt"
 set +e
 LATENT_TIMES="$root/kp-$label-times.jsonl" LATENT_FULL=0 LATENT_ASSERT_NO_OUTPUT=1 python3 "$root/kp-time.py" "$root/kp-$label-rss.json" "$destination/latent-census" "$root/kp-adapted/src/compiler" "$root/kp-$label-latent.jsonl" > "$root/kp-$label-latent.log" 2>&1
 status=$?
 set -e
 echo "$status" > "$root/kp-$label-exit.txt"
 cat /sys/fs/cgroup/memory.events > "$root/kp-$label-memory-after.txt"
 [[ $status == 0 ]] || exit "$status"
done
