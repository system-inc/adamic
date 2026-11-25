#!/bin/bash
# #42q54q5 before review, sealed: the per-call allocation probe, count.h's buffers and adamic_count_snapshot. Vet, the
# new oracle test verbose, both mutants planted by python and restored with a check, then internal/native_runtime,
# internal/native, internal/native/runtimeproducts and internal/leakcheck whole, internal/oracle's count, leak,
# record and region tests, and internal/oracle whole. Never lands.
#
# It runs after the products are in place: sandboxed_repositories && sandboxed_npm stage3/api &&
# sandboxed_products ./internal/oracle ./internal/native
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/42q54q5"
mkdir -p "$T"
COUNT=internal/native_runtime/c/count.h
ARRAY=internal/native_runtime/c/array.c
cp "$COUNT" "$T/count.orig"
cp "$ARRAY" "$T/array.orig"

go vet ./internal/native_runtime ./internal/native ./internal/leakcheck ./internal/oracle > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }

WINDOW=TestCountWindow
go test -count=1 -v -run "^$WINDOW\$" ./internal/oracle > "$T/window.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/window.log"
grep -E -- '^\s+count_window_test.go:[0-9]+: (allocations|frees|buffers) ' "$T/window.log"
[ $code = 0 ] && echo "PASS the count window pins" || { grep -E '_test.go:[0-9]+:' -A4 "$T/window.log"; tail -60 "$T/window.log"; echo "FAIL count window pins"; exit 1; }

# plant FILE OLD NEW: replaces OLD, found exactly once, with NEW.
plant() {
	python3 -I - "$1" "$2" "$3" <<'PY'
import sys
path, old, new = sys.argv[1], sys.argv[2], sys.argv[3]
text = open(path).read()
if text.count(old) != 1:
    sys.exit("seam not found once: " + old)
open(path, "w").write(text.replace(old, new, 1))
PY
}
# restore FILE ORIGINAL: puts the file back and checks the tree holds no change to it.
restore() {
	cp "$2" "$1"
	git diff --quiet -- "$1" || { echo "FAIL $1 not restored"; exit 1; }
}
# red LOG NAME...: how many of the named subtests failed in LOG.
red() {
	local log=$1 count=0
	shift
	for name in "$@"; do
		grep -qE -- "--- FAIL: $name( |$)" "$log" && count=$((count + 1))
	done
	echo $count
}
green() {
	local log=$1
	shift
	for name in "$@"; do
		grep -qE -- "--- FAIL: $name( |$)" "$log" && return 1
	done
	return 0
}

# M1: ADAMIC_COUNT_BUFFER counts nothing under ADAMIC_COUNT. The push, map and made before rows go red, each on its
# buffers alone, reading 0; the objects and arithmetic rows, whose buffers are 0 either way, stay green.
plant "$COUNT" '#define ADAMIC_COUNT_BUFFER() atomic_fetch_add_explicit(&adamic_counted.buffers, 1, memory_order_relaxed)' '#define ADAMIC_COUNT_BUFFER() ((void)0)' || { restore "$COUNT" "$T/count.orig"; echo "FAIL M1 seam"; exit 1; }
go test -count=1 -v -run "^$WINDOW\$" ./internal/oracle > "$T/m1.log" 2>&1
restore "$COUNT" "$T/count.orig"
n=$(red "$T/m1.log" "$WINDOW/push" "$WINDOW/map" "$WINDOW/made_before")
buffers=$(grep -cE '^\s+count_window_test.go:[0-9]+: buffers across the call: 0, want [0-9]+$' "$T/m1.log")
if [ "$n" = 3 ] && [ "$buffers" = 3 ] && green "$T/m1.log" "$WINDOW/objects" "$WINDOW/arithmetic"; then
	echo "PASS M1: the push, map and made before rows red on their buffers, the rest green"
else
	grep -E -- '^\s*--- (PASS|FAIL)' "$T/m1.log"; tail -40 "$T/m1.log"; echo "FAIL M1: $n of 3 red, $buffers on buffers"; exit 1
fi

# M2, the must's own: push's first allocation skipped again, its count guarded by the capacity it had. The made before
# row goes red reading 0 buffers and the push row reading 3; the map, objects and arithmetic rows stay green.
plant "$ARRAY" $'\t\t// storage on the first push inside the body (count.h).\n\t\tADAMIC_COUNT_BUFFER();\n' $'\t\t// storage on the first push inside the body (count.h).\n\t\tif (array->capacity != 0) {\n\t\t\tADAMIC_COUNT_BUFFER();\n\t\t}\n' || { restore "$ARRAY" "$T/array.orig"; echo "FAIL M2 seam"; exit 1; }
go test -count=1 -v -run "^$WINDOW\$" ./internal/oracle > "$T/m2.log" 2>&1
restore "$ARRAY" "$T/array.orig"
n=$(red "$T/m2.log" "$WINDOW/made_before" "$WINDOW/push")
first=$(grep -cE '^\s+count_window_test.go:[0-9]+: buffers across the call: 0, want 1$' "$T/m2.log")
pushed=$(grep -cE '^\s+count_window_test.go:[0-9]+: buffers across the call: 3, want 4$' "$T/m2.log")
if [ "$n" = 2 ] && [ "$first" = 1 ] && [ "$pushed" = 1 ] && green "$T/m2.log" "$WINDOW/map" "$WINDOW/objects" "$WINDOW/arithmetic"; then
	echo "PASS M2: the made before and push rows red on their buffers, the rest green"
else
	grep -E -- '^\s*--- (PASS|FAIL)' "$T/m2.log"; tail -40 "$T/m2.log"; echo "FAIL M2: $n of 2 red, made before $first and push $pushed on buffers"; exit 1
fi

go test -count=1 -timeout 0 ./internal/native_runtime ./internal/native ./internal/native/runtimeproducts ./internal/leakcheck > "$T/whole.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/whole.log"
[ $code = 0 ] && echo "PASS native_runtime, native, runtimeproducts and leakcheck whole" || { grep -B3 -A20 -- '--- FAIL' "$T/whole.log" | head -100; echo "FAIL whole"; exit 1; }

go test -count=1 -timeout 0 -run 'Count|Leak|Record|Region' ./internal/oracle > "$T/oracle-counts.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/oracle-counts.log"
[ $code = 0 ] && echo "PASS oracle's count, leak, record and region tests" || { grep -B3 -A20 -- '--- FAIL' "$T/oracle-counts.log" | head -100; echo "FAIL oracle counts"; exit 1; }

go test -count=1 -timeout 0 ./internal/oracle > "$T/oracle.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/oracle.log"
[ $code = 0 ] && echo "PASS oracle whole" || { grep -B3 -A20 -- '--- FAIL' "$T/oracle.log" | head -100; echo "FAIL oracle whole"; exit 1; }
echo VERDICT PASS
