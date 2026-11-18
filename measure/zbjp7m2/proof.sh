#!/usr/bin/env bash
# #zbjp7m2's per-thread cursor: the parallel and string tests of internal/native, the two new mutants and the cache
# race mutant, and gofmt. Sealed. Never lands.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
gofmt -l internal/native/parallel_test.go | grep . && { echo "FAIL gofmt"; exit 1; } || echo "PASS gofmt"
sealed_products ./internal/native > "$TMPDIR/products.log" 2>&1 && echo "PASS products" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
go test -count=1 -timeout 0 -v -run 'TestParallel|TestString' ./internal/native/ > "$TMPDIR/tests.log" 2>&1; code=$?
grep -E -- '^(--- |ok |FAIL|PASS)|caught:|shared cursors' "$TMPDIR/tests.log" | grep -vE -- '--- PASS: .*/' | head -80
[ $code = 0 ] && echo "PASS tests" || { echo "FAIL tests"; grep -B2 -A20 -- '--- FAIL' "$TMPDIR/tests.log" | head -80; exit 1; }
