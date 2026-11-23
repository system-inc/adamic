#!/bin/bash
# #2v3xtgm wave 1 before review, sealed: vet, the cast-cause tests verbose, mutant M3 (castExtendsArray false: the two
# array pins red, the dictionary pin green), internal/lower and internal/dry whole, and both censuses. Never lands.
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/wave1"
mkdir -p "$T"
F=internal/lower/cast_stop.go
cp "$F" "$T/cast_stop.go.orig"
go vet ./internal/lower > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }
go test -count=1 -v -run '^TestCastStopNamesItsCause$|^TestUncheckableCastsStayRefused$' ./internal/lower > "$T/pins.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/pins.log"
[ $code = 0 ] && echo "PASS the cast-cause pins" || { tail -40 "$T/pins.log"; echo "FAIL pins"; exit 1; }

python3 -I - "$F" <<'PY' || { cp "$T/cast_stop.go.orig" "$F"; echo "FAIL mutant seam"; exit 1; }
import sys
path = sys.argv[1]
text = open(path).read()
seam = "func (l *lowering) castExtendsArray(proven *checker.Type) bool {\n"
if text.count(seam) != 1:
    sys.exit("seam not found once")
open(path, "w").write(text.replace(seam, seam + "\treturn false\n", 1))
PY
go test -count=1 -v -run '^TestCastStopNamesItsCause$' ./internal/lower > "$T/m3.log" 2>&1
cp "$T/cast_stop.go.orig" "$F"
git diff --quiet -- "$F" || { echo "FAIL cast_stop.go not restored"; exit 1; }
grep -E -- '--- FAIL' "$T/m3.log"
red=$(grep -cE -- '--- FAIL: TestCastStopNamesItsCause/(missing_list|node_or_node_array)' "$T/m3.log")
if [ "$red" = 2 ] && ! grep -qE -- '--- FAIL: TestCastStopNamesItsCause/map-like_dictionary' "$T/m3.log"; then
	echo "PASS M3: the two array pins red, the dictionary pin green"
else
	grep -c "unreachable code\|declared and not used" "$T/m3.log"; tail -30 "$T/m3.log"; echo "FAIL M3: $red of the two array pins red"; exit 1
fi

go test -count=1 -timeout 0 ./internal/lower ./internal/dry ./internal/compilercensus ./cmd/adamic-gate/parallelcensus > "$T/whole.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/whole.log"
[ $code = 0 ] && echo "PASS lower, dry and both censuses whole" || { grep -B3 -A20 -- '--- FAIL' "$T/whole.log" | head -100; echo "FAIL whole"; exit 1; }
echo VERDICT PASS
