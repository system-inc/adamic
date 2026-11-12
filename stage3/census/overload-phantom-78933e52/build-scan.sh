#!/bin/bash
set -eu
source /workspace/adamic-tools/env.sh
label=$1
scripts=/workspace/op2-report/stage3/census/overload-phantom-78933e52
build=/workspace/oph-build-$label
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
cd /workspace/oph-$label
run scan-build go build -buildvcs=false -overlay="$build/scan-overlay/overlay.json" -o "$build/entry-census-scan" ./stage3/census/latent/tool
