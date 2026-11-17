#!/usr/bin/env bash
# proof.sh: #snypn8c c5's proof, a scratch script that never lands. Run from the tree's root after the sealed preamble.
# PART=pins runs gofmt, the pins and the two mutants; PART=whole adds internal/lower and internal/ir whole. Verdict
# lines begin "c5: verdict".
set -u
root=$PWD
out="${TMPDIR:-/tmp}/c5-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$out"
say() { echo "$*" | tee -a "$out/verdicts.log"; }
pins='^(TestTypeParameterSourceIntoWiderArrayStillRefuses|TestTypeParameterSourceIntoWholeEnumSlotLowers|TestTypeParameterSourceIsReadAtItsConstraint|TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint|TestTypeParameterSourceConstrainedToTheMemberLowers|TestUnconstrainedTypeParameterSourceIsJudgedAsBefore)$'

unformatted=$(gofmt -l internal/lower/invariance.go internal/lower/enum_type_parameter_source_test.go)
[ -z "$unformatted" ] && say "c5: verdict gofmt PASS" || say "c5: verdict gofmt FAIL: $unformatted"

go test -count=1 -timeout 30m -run "$pins" -v ./internal/lower/ > "$out/pins.log" 2>&1
code=$?
grep -E '^\s*--- (PASS|FAIL)' "$out/pins.log" | cut -c1-200 | head -20
[ "$code" = 0 ] && say "c5: verdict pins PASS" || { say "c5: verdict pins FAIL: exit $code"; grep -E '^\s+enum_type_parameter_source_test.go' "$out/pins.log" | cut -c1-400 | head -10; }

# mutant <name> <file> <old> <new> <tests> <must>: the one exact edit applied, the tests run, the file restored
# from a copy. RED only when every test or subtest named in must (a | list of names as go prints them) failed.
mutant() {
	local name=$1 file=$2 old=$3 new=$4 tests=$5 must=$6
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
		say "c5: verdict mutant $name BROKEN: the edit didn't apply"
		cp "$out/$name.orig" "$file"
		return
	fi
	go test -count=1 -timeout 30m -run "$tests" -v ./internal/lower/ > "$out/$name.log" 2>&1
	local code=$?
	cp "$out/$name.orig" "$file"
	grep -E '^\s*--- FAIL' "$out/$name.log" | cut -c1-200 | head -10
	if grep -q 'build failed\|\[setup failed\]\|cannot use\|undefined:' "$out/$name.log"; then
		say "c5: verdict mutant $name BROKEN: it didn't compile"
		tail -n 20 "$out/$name.log"
		return
	fi
	local missing=""
	local IFS='|'
	for test in $must; do
		grep -qE "^\s*--- FAIL: $test " "$out/$name.log" || missing="$missing $test"
	done
	if [ "$code" != 0 ] && [ -z "$missing" ]; then
		say "c5: verdict mutant $name RED: $(grep -cE '^\s*--- FAIL' "$out/$name.log") failing"
	else
		say "c5: verdict mutant $name SURVIVED: exit $code, not failed:$missing"
	fi
}
mutant m1-no-read internal/lower/invariance.go \
	'	if from.Flags()&checker.TypeFlagsTypeParameter != 0 && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool { return member.Flags()&checker.TypeFlagsTypeParameter != 0 }) {' \
	'	if false && from.Flags()&checker.TypeFlagsTypeParameter != 0 && !slices.ContainsFunc(l.definedMembers(to), func(member *checker.Type) bool { return member.Flags()&checker.TypeFlagsTypeParameter != 0 }) {' \
	'^(TestTypeParameterSourceIntoWholeEnumSlotLowers|TestTypeParameterSourceIsReadAtItsConstraint)$' \
	'TestTypeParameterSourceIntoWholeEnumSlotLowers|TestTypeParameterSourceIsReadAtItsConstraint'
mutant m2-passes internal/lower/invariance.go \
	'			return l.widened(constraint, to, visited)' \
	'			return nil' \
	'^(TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint|TestTypeParameterSourceIntoWiderArrayStillRefuses)$' \
	'TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint/constrained_to_number|TestTypeParameterSourceIntoMemberSlotIsRefusedAsItsConstraint/a_chain_to_number'
git -C "$root" diff --stat -- internal/lower 2> /dev/null | tail -n 3

if [ "${PART:-pins}" = whole ]; then
	for package in lower ir; do
		go test -count=1 -timeout 40m ./internal/$package/ > "$out/whole-$package.log" 2>&1
		code=$?
		grep -E '^(--- FAIL|FAIL|ok|panic)' "$out/whole-$package.log" | cut -c1-300 | head -40
		[ "$code" = 0 ] && say "c5: verdict $package whole PASS" || say "c5: verdict $package whole FAIL: exit $code"
	done
fi
echo "c5: summary"
cat "$out/verdicts.log"
grep -q 'FAIL\|SURVIVED\|BROKEN' "$out/verdicts.log" && exit 1
exit 0
