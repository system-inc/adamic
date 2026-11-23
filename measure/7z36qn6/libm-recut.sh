#!/bin/bash
# #7z36qn6: the libm head re-cut with MathLinkFlags in leanRuntimeEntries. Vet, the libm pin, and every package that
# reads the change whole: internal/native, internal/native_runtime and internal/native/runtimeproducts. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/libm-recut"
mkdir -p "$T"
go vet ./internal/native ./internal/native_runtime ./internal/native/runtimeproducts ./internal/oracle > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }
go test -count=1 -v -run '^TestMathLinksLibmOnlyWhereCalled$|^TestTheLeanBuildReachesOnlyItsDeclaredFiles$' ./internal/oracle ./internal/native/runtimeproducts > "$T/pins.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/pins.log"
[ $code = 0 ] && echo "PASS the pins" || { grep -E '_test.go:[0-9]+:' -A4 "$T/pins.log"; tail -40 "$T/pins.log"; echo "FAIL pins"; exit 1; }
go test -count=1 -timeout 0 ./internal/native ./internal/native_runtime ./internal/native/runtimeproducts > "$T/whole.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/whole.log"
[ $code = 0 ] && echo "PASS native, native_runtime and runtimeproducts whole" || { grep -B3 -A20 -- '--- FAIL' "$T/whole.log" | head -100; echo "FAIL whole"; exit 1; }
echo VERDICT PASS
