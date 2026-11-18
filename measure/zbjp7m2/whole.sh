#!/usr/bin/env bash
# #zbjp7m2 before Loom: one package whole (the first argument), sealed, with its products settled; for internal/native it also
# prints what the statics scanner sees in string_index.c. Never lands.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sealed-job.sh || { echo "FAIL sealed environment"; exit 1; }
package=$1
sealed_products "$package" > "$TMPDIR/products.log" 2>&1 && echo "PASS products $package" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
if [ "$package" = ./internal/native ]; then
	( cd internal/native && go test -count=1 -run '^TestRuntimeStaticsAreListed$' -runtime-statics-list . 2>&1 | grep 'string_index' )
fi
go test -count=1 -timeout 0 "$package" > "$TMPDIR/tests.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$TMPDIR/tests.log" | head -40
[ $code = 0 ] && echo "PASS $package whole" || { grep -B3 -A25 -- '--- FAIL' "$TMPDIR/tests.log" | head -120; echo "FAIL $package"; exit 1; }
