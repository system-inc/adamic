#!/usr/bin/env bash
# proof.sh: #y1mw6wj's proof, a scratch script that never lands. Run from the tree's root after the sealed preamble.
# PART=pins runs gofmt, the three pins and the two mutants; PART=whole adds internal/lower whole. Verdict lines begin
# "y1: verdict".
set -u
root=$PWD
out="${TMPDIR:-/tmp}/y1-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$out"
say() { echo "$*" | tee -a "$out/verdicts.log"; }
pins='^(TestAndLeftFalsyArms|TestAndLeftObjectArmsLower|TestAndLeftKeepsPrimitiveArms)$'

unformatted=$(gofmt -l internal/lower/logical_and_left.go internal/lower/logical_and_left_test.go internal/lower/invariance.go)
[ -z "$unformatted" ] && say "y1: verdict gofmt PASS" || say "y1: verdict gofmt FAIL: $unformatted"

go test -count=1 -timeout 30m -run "$pins" -v ./internal/lower/ > "$out/pins.log" 2>&1
code=$?
grep -E '^(=== RUN|--- (PASS|FAIL)|\s+logical_and_left_test.go)' "$out/pins.log" | head -60
[ "$code" = 0 ] && say "y1: verdict pins PASS" || { say "y1: verdict pins FAIL: exit $code"; tail -n 40 "$out/pins.log"; }

# mutant <name> <file> <old> <new> <tests>: the one exact edit applied, the tests run, the file restored from a copy.
mutant() {
	local name=$1 file=$2 old=$3 new=$4 tests=$5
	cp "$file" "$out/$name.orig"
	python3 -I - "$file" "$old" "$new" << 'PYTHON'
import sys
path, old, new = sys.argv[1:4]
text = open(path).read()
if text.count(old) != 1:
    sys.exit(f'mutant: {old!r} occurs {text.count(old)} times in {path}')
open(path, 'w').write(text.replace(old, new))
PYTHON
	if [ $? != 0 ]; then
		say "y1: verdict mutant $name BROKEN: the edit didn't apply"
		cp "$out/$name.orig" "$file"
		return
	fi
	go test -count=1 -timeout 30m -run "$tests" -v ./internal/lower/ > "$out/$name.log" 2>&1
	local code=$?
	cp "$out/$name.orig" "$file"
	grep -E '^(--- FAIL|\s+logical_and_left_test.go)' "$out/$name.log" | cut -c1-400 | head -30
	if grep -q 'build failed\|\[setup failed\]\|cannot use\|undefined:' "$out/$name.log"; then
		say "y1: verdict mutant $name BROKEN: it didn't compile"
		tail -n 20 "$out/$name.log"
	elif [ "$code" != 0 ] && grep -q '^--- FAIL' "$out/$name.log"; then
		say "y1: verdict mutant $name RED: $(grep -c '^\s*--- FAIL' "$out/$name.log") failing"
	else
		say "y1: verdict mutant $name SURVIVED: exit $code"
	fi
}
mutant m1-no-drop internal/lower/invariance.go '		if andLeftOperand(node) {' '		if false && andLeftOperand(node) {' '^TestAndLeftObjectArmsLower$'
mutant m2-no-string internal/lower/logical_and_left.go '[]*checker.Type{l.checker.GetStringType(), l.checker.GetNumberType(), l.checker.GetBooleanType(), l.checker.GetBigIntType(), l.checker.GetESSymbolType()}' '[]*checker.Type{l.checker.GetNumberType(), l.checker.GetBooleanType(), l.checker.GetBigIntType(), l.checker.GetESSymbolType()}' '^(TestAndLeftFalsyArms|TestAndLeftKeepsPrimitiveArms)$'
mutant m3-optional-required internal/lower/logical_and_left.go 'if property.Flags&ast.SymbolFlagsOptional == 0 && !l.primitiveHas(property.Name) {' 'if !l.primitiveHas(property.Name) {' '^TestAndLeftFalsyArms$'
git -C "$root" diff --stat -- internal/lower 2> /dev/null | tail -n 3

if [ "${PART:-pins}" = whole ]; then
	go test -count=1 -timeout 40m ./internal/lower/ > "$out/whole.log" 2>&1
	code=$?
	grep -E '^(--- FAIL|FAIL|ok|panic)' "$out/whole.log" | cut -c1-300 | head -60
	[ "$code" = 0 ] && say "y1: verdict lower whole PASS" || say "y1: verdict lower whole FAIL: exit $code"
fi
echo "y1: summary"
cat "$out/verdicts.log"
grep -q 'FAIL\|SURVIVED\|BROKEN' "$out/verdicts.log" && exit 1
exit 0
