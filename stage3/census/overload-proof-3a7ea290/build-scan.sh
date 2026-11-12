#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
label=$1
scripts=/workspace/op2-report/stage3/census/overload-proof-3a7ea290
build=/workspace/op2-build-$label
run() {
 local name=$1; shift
 set +e
 python3 "$scripts/run.py" "$build/$name.log" "$build/$name.time.json" "$@"
 local rc=$?
 set -e
 echo "$name child_exit=$rc"
 test "$rc" = 0 || exit "$rc"
}
run scan-overlay python3 "$scripts/scan_overlay.py" "$build/entry-overlay" "$build/scan-overlay"
run scan-format gofmt -w "$build/scan-overlay/units.go"
cd /workspace/op2-$label
run scan-build go build -buildvcs=false -overlay="$build/scan-overlay/overlay.json" -o "$build/entry-census-scan" ./stage3/census/latent/tool
