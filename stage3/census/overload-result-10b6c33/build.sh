#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
label=$1
src=/workspace/overload-$label
out=/workspace/overload-build-$label
mkdir -p "$out"
run() {
 local name=$1; shift
 set +e
 python3 /workspace/overload-run.py "$out/$name.log" "$out/$name.time.json" "$@"
 local rc=$?
 set -e
 echo "$name child_exit=$rc"
 test "$rc" = 0 || exit "$rc"
}
cd "$src"
run overlay python3 /workspace/overload-parent/stage3/census/latent/make_overlay.py "$src" "$out/latent-overlay"
run format gofmt -w "$out/latent-overlay/"*.go
run entry-overlay python3 /workspace/overload-parent/stage3/meter/entry_overlay.py "$out/latent-overlay" "$out/entry-overlay"
run entry-format gofmt -w "$out/entry-overlay/"*.go
run entry-build go build -buildvcs=false -overlay="$out/entry-overlay/overlay.json" -o "$out/entry-census" ./stage3/census/latent/tool
run checker-build go build -o "$out/census" ./stage3/census/tool
