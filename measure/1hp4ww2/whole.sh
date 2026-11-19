#!/usr/bin/env bash
# #1hp4ww2's cached for-in key list before Loom, sealed. The first argument is a package to run whole (products
# settled, stage3/api installed for internal/oracle), or m1 or m2: the insert's or define_hidden's let-go
# removed must turn TestRecordForInCachedKeys red, uncached. Never lands.
set -uo pipefail
source /opt/adamic-block/environment.sh && source cloud/sandboxed-job.sh || { echo "FAIL sandboxed environment"; exit 1; }
target=$1
case "$target" in
m1|m2)
	sandboxed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 && echo "PASS products ./internal/oracle" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
	run() { ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 0 -run '^TestRecordForInCachedKeys$' ./internal/oracle > "$TMPDIR/$1.log" 2>&1; }
	run clean && echo "PASS the pin is green on the head" || { tail -40 "$TMPDIR/clean.log"; echo "FAIL the pin is red on the head"; exit 1; }
	python3 -I - "$target" <<'PY' || { echo "FAIL the mutant didn't apply"; exit 1; }
import sys
which = sys.argv[1]
if which == "m1":
    path = "internal/native_runtime/c/map.c"
    old = "\tadamic_map_forget_key_list(map);\n\tif (map->used == map->capacity) {"
    new = "\tif (map->used == map->capacity) {"
else:
    path = "internal/native_runtime/c/record.c"
    old = "\t\tadamic_map_forget_key_list(map);\n\t\tadamic_map_entry *entry = entry_of(slot);"
    new = "\t\tadamic_map_entry *entry = entry_of(slot);"
text = open(path).read()
assert text.count(old) == 1, "the let-go isn't where the head put it"
open(path, "w").write(text.replace(old, new))
PY
	echo "mutant $target applied"
	if run mutant; then echo "FAIL mutant $target survived: TestRecordForInCachedKeys stayed green"; exit 1; fi
	grep -E -- '--- FAIL|differ|want|got' "$TMPDIR/mutant.log" | head -10 | cut -c1-300
	echo "PASS mutant $target is caught by TestRecordForInCachedKeys"
	exit 0 ;;
esac
[ "$target" = ./internal/oracle ] && { sandboxed_npm stage3/api > "$TMPDIR/npm.log" 2>&1 && echo "PASS npm stage3/api" || { tail -20 "$TMPDIR/npm.log"; echo "FAIL npm"; exit 1; }; }
sandboxed_products "$target" > "$TMPDIR/products.log" 2>&1 && echo "PASS products $target" || { tail -30 "$TMPDIR/products.log"; echo "FAIL products"; exit 1; }
go test -count=1 -timeout 0 "$target" > "$TMPDIR/tests.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$TMPDIR/tests.log" | head -40
[ $code = 0 ] && echo "PASS $target whole" || { grep -B3 -A25 -- '--- FAIL' "$TMPDIR/tests.log" | head -150; echo "FAIL $target"; exit 1; }
