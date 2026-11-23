#!/bin/bash
# #2v3xtgm wave 2 before review, sealed: the checked dictionary view, object as MapLike<unknown>. Vet, the new and old
# cast pins verbose, the new oracle test verbose, mutants M4 to M7 each planted by python and restored with a check,
# then internal/lower, internal/ir, internal/native, internal/javascript and cmd/adamic-admission-delta whole and
# internal/oracle's records and cast tests. Never lands.
#
# It runs after the oracle's products are in place: sandboxed_repositories && sandboxed_npm stage3/api &&
# sandboxed_products ./internal/oracle
set -uo pipefail
echo "box $(hostname) $(nproc) cpus, load $(cut -d' ' -f1-3 /proc/loadavg), tree $(git rev-parse HEAD)"
T="$TMPDIR/wave2"
mkdir -p "$T"
VIEW=internal/lower/cast_dictionary_view.go
MARK=internal/javascript/records.go
cp "$VIEW" "$T/view.orig"
cp "$MARK" "$T/mark.orig"

go vet ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./internal/oracle > "$T/vet.log" 2>&1 && echo "PASS vet" || { cat "$T/vet.log"; echo "FAIL vet"; exit 1; }

PINS='^TestDictionaryViewCasts$|^TestDictionaryViewKeepsAFixedSourceStopped$|^TestDictionaryViewRefusedInAdamic$|^TestCastStopNamesItsCause$|^TestUncheckableCastsStayRefused$'
go test -count=1 -v -run "$PINS" ./internal/lower > "$T/pins.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/pins.log"
[ $code = 0 ] && echo "PASS the lower pins" || { tail -40 "$T/pins.log"; echo "FAIL lower pins"; exit 1; }

ORACLE='^TestDictionaryViewAgreesWithNode$|^TestDictionaryViewStopsAtTheCast$'
go test -count=1 -v -run "$ORACLE" ./internal/oracle > "$T/oracle.log" 2>&1; code=$?
grep -E -- '^\s*--- (PASS|FAIL)' "$T/oracle.log"
[ $code = 0 ] && echo "PASS the oracle pins" || { tail -60 "$T/oracle.log"; echo "FAIL oracle pins"; exit 1; }

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
STOPS=TestDictionaryViewStopsAtTheCast

# M4: an unchecked narrow at the cast. The array, literal and class rows go red.
plant "$VIEW" "ir.Narrow{Value: held, To: ir.Record, Checked: true, Message: ir.DictionaryViewRecordStop, ReferencesMessage: ir.DictionaryViewReferencesStop}" "ir.Narrow{Value: held, To: ir.Record}" || { restore "$VIEW" "$T/view.orig"; echo "FAIL M4 seam"; exit 1; }
go test -count=1 -v -run "^$STOPS\$" ./internal/oracle > "$T/m4.log" 2>&1
restore "$VIEW" "$T/view.orig"
n=$(red "$T/m4.log" "$STOPS/array_cast.ts" "$STOPS/fixed_literal.ts" "$STOPS/class_instance.ts")
[ "$n" = 3 ] && echo "PASS M4: the array, literal and class rows red" || { tail -40 "$T/m4.log"; echo "FAIL M4: $n of 3 red"; exit 1; }

# M5: the references test dropped. The number and boolean rows go red; the array, literal and class rows stay green.
plant "$VIEW" ", ReferencesMessage: ir.DictionaryViewReferencesStop}" "}" || { restore "$VIEW" "$T/view.orig"; echo "FAIL M5 seam"; exit 1; }
go test -count=1 -v -run "^$STOPS\$" ./internal/oracle > "$T/m5.log" 2>&1
restore "$VIEW" "$T/view.orig"
n=$(red "$T/m5.log" "$STOPS/number_record.ts" "$STOPS/boolean_record.ts")
if [ "$n" = 2 ] && green "$T/m5.log" "$STOPS/array_cast.ts" "$STOPS/fixed_literal.ts" "$STOPS/class_instance.ts"; then
	echo "PASS M5: the number and boolean rows red, the rest green"
else
	tail -40 "$T/m5.log"; echo "FAIL M5: $n of 2 red"; exit 1
fi

# M6: the JavaScript mark always "reference". The number and boolean rows go red on the javascript backend.
plant "$MARK" 'quote(recordValueKind(r.Element))' 'quote("reference")' || { restore "$MARK" "$T/mark.orig"; echo "FAIL M6 seam"; exit 1; }
go test -count=1 -v -run "^$STOPS\$" ./internal/oracle > "$T/m6.log" 2>&1
restore "$MARK" "$T/mark.orig"
n=$(red "$T/m6.log" "$STOPS/number_record.ts" "$STOPS/boolean_record.ts")
javascript=$(grep -cE '^\s+records_dictionary_view_test.go:[0-9]+: javascript: ' "$T/m6.log")
if [ "$n" = 2 ] && [ "$javascript" = 2 ]; then
	echo "PASS M6: the number and boolean rows red on javascript"
else
	tail -40 "$T/m6.log"; echo "FAIL M6: $n of 2 red, $javascript on javascript"; exit 1
fi

# M7: the write walk dropped. The written and deleted rows go red; every other row stays green.
plant "$VIEW" $'func (l *lowering) dictionaryViewWritten(present *checker.Type) bool {\n' $'func (l *lowering) dictionaryViewWritten(present *checker.Type) bool {\n\treturn false\n' || { restore "$VIEW" "$T/view.orig"; echo "FAIL M7 seam"; exit 1; }
go test -count=1 -v -run '^TestDictionaryViewCasts$' ./internal/lower > "$T/m7.log" 2>&1
restore "$VIEW" "$T/view.orig"
CASTS=TestDictionaryViewCasts
n=$(red "$T/m7.log" "$CASTS/written_through_a_dictionary" "$CASTS/deleted_through_a_dictionary")
if [ "$n" = 2 ] && green "$T/m7.log" "$CASTS/maplike_of_unknown" "$CASTS/undefined_dropped" "$CASTS/number_values" "$CASTS/union-held_values"; then
	echo "PASS M7: the written and deleted rows red, the rest green"
else
	tail -40 "$T/m7.log"; echo "FAIL M7: $n of 2 red"; exit 1
fi

go test -count=1 -timeout 0 ./internal/lower ./internal/ir ./internal/native ./internal/javascript ./cmd/adamic-admission-delta ./stage3/census/latent/statecopy > "$T/whole.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/whole.log"
[ $code = 0 ] && echo "PASS lower, ir, native, javascript and the admission delta whole" || { grep -B3 -A20 -- '--- FAIL' "$T/whole.log" | head -100; echo "FAIL whole"; exit 1; }

go test -count=1 -timeout 0 -run 'DictionaryView|OptionsBase|Record|Cast' ./internal/oracle > "$T/oracle-records.log" 2>&1; code=$?
grep -E -- '^(--- FAIL|ok |FAIL)' "$T/oracle-records.log"
[ $code = 0 ] && echo "PASS oracle's records and cast tests" || { grep -B3 -A20 -- '--- FAIL' "$T/oracle-records.log" | head -100; echo "FAIL oracle records"; exit 1; }
echo VERDICT PASS
