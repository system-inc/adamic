#!/bin/bash
# #42q54q5 before review, sealed: the per-call allocation probe, count.h's growths and adamic_count_snapshot. Vet, the
# new oracle test verbose, the mutant planted by python and restored with a check, then internal/native_runtime,
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
cp "$COUNT" "$T/count.orig"

go vet ./internal/native_runtime ./internal/native ./internal/leakcheck ./internal/oracle > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }

WINDOW=TestCountWindow
go test -count=1 -v -run "^$WINDOW\$" ./internal/oracle > "$T/window.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/window.log"
grep -E -- '^\s+count_window_test.go:[0-9]+: (allocations|frees|growths) ' "$T/window.log"
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

# The mutant: ADAMIC_COUNT_GROWTH counts nothing under ADAMIC_COUNT. The push and map rows go red, each on its growths
# alone; the objects and arithmetic rows, whose growths are 0 either way, stay green.
plant "$COUNT" '#define ADAMIC_COUNT_GROWTH() atomic_fetch_add_explicit(&adamic_counted.growths, 1, memory_order_relaxed)' '#define ADAMIC_COUNT_GROWTH() ((void)0)' || { restore "$COUNT" "$T/count.orig"; echo "FAIL mutant seam"; exit 1; }
go test -count=1 -v -run "^$WINDOW\$" ./internal/oracle > "$T/mutant.log" 2>&1
restore "$COUNT" "$T/count.orig"
n=$(red "$T/mutant.log" "$WINDOW/push" "$WINDOW/map")
growths=$(grep -cE '^\s+count_window_test.go:[0-9]+: growths across the call: 0, want [0-9]+$' "$T/mutant.log")
if [ "$n" = 2 ] && [ "$growths" = 2 ] && green "$T/mutant.log" "$WINDOW/objects" "$WINDOW/arithmetic"; then
	echo "PASS mutant: the push and map rows red on their growths, the rest green"
else
	grep -E -- '^\s*--- (PASS|FAIL)' "$T/mutant.log"; tail -40 "$T/mutant.log"; echo "FAIL mutant: $n of 2 red, $growths on growths"; exit 1
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
