#!/usr/bin/env bash
# #2ccxtyr's inline adamic_array_set before Loom, sealed. The first argument is a package to run whole
# (with its products settled, and stage3/api installed for internal/oracle), or "mutant": the inline
# write without its bounds test must turn TestStage3EnumSparseArrayBoundary red, uncached. Never lands.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh || { echo "FAIL sandboxed environment"; exit 1; }
target=$1
if [ "$target" = mutant ]; then
	sandboxed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 && echo "PASS products ./internal/oracle" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
	run() { ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 0 -run '^TestStage3EnumSparseArrayBoundary$' ./internal/oracle > "$TMPDIR/$1.log" 2>&1; }
	run clean && echo "PASS the pin is green on the head" || { tail -40 "$TMPDIR/clean.log"; echo "FAIL the pin is red on the head"; exit 1; }
	python3 -I - internal/native_runtime/c/adamic.h <<'PY' || { echo "FAIL the mutant didn't apply"; exit 1; }
import sys
path = sys.argv[1]; text = open(path).read()
old = "\tif (slot == NULL) {\n\t\tadamic_array_set_outside(array, index);\n\t}\n\tif (array->references) {\n\t\tadamic_array_set_held("
new = "\tif (array->references) {\n\t\tadamic_array_set_held("
assert text.count(old) == 1, "the bounds test isn't where the head put it"
open(path, "w").write(text.replace(old, new))
PY
	echo "mutant applied: the inline adamic_array_set without its slot == NULL test"
	if run mutant; then echo "FAIL the mutant survived: TestStage3EnumSparseArrayBoundary stayed green"; exit 1; fi
	grep -E -- '--- FAIL|exit|differ' "$TMPDIR/mutant.log" | head -8
	echo "PASS the mutant is caught by TestStage3EnumSparseArrayBoundary"
	exit 0
fi
[ "$target" = ./internal/oracle ] && { sandboxed_npm stage3/api > "$TMPDIR/npm.log" 2>&1 && echo "PASS npm stage3/api" || { tail -20 "$TMPDIR/npm.log"; echo "FAIL npm"; exit 1; }; }
sandboxed_products "$target" > "$TMPDIR/products.log" 2>&1 && echo "PASS products $target" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
go test -count=1 -timeout 0 "$target" > "$TMPDIR/tests.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$TMPDIR/tests.log" | head -40
[ $code = 0 ] && echo "PASS $target whole" || { grep -B3 -A25 -- '--- FAIL' "$TMPDIR/tests.log" | head -120; echo "FAIL $target"; exit 1; }
