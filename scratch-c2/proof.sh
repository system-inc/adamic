#!/usr/bin/env bash
# proof.sh: #snypn8c c2's proof, a scratch script that never lands. Run from the tree's root. Verdict lines begin
# "c2: verdict"; it exits 1 on any FAIL, SURVIVED or BROKEN.
set -u
source /opt/adamic-block/environment.sh
source cloud/sealed-job.sh
root=$PWD
out="${TMPDIR:-/tmp}/c2-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$out"
say() { echo "$*" | tee -a "$out/verdicts.log"; }

sealed_products ./internal/lower ./internal/ir > "$out/products.log" 2>&1
code=$?
[ "$code" = 0 ] && say "c2: verdict products PASS" || { say "c2: verdict products FAIL: exit $code"; tail -n 20 "$out/products.log"; }

edited='internal/lower/enums.go internal/lower/invariance.go internal/lower/enum_slot_part_test.go'
unformatted=$(gofmt -l $edited)
[ -z "$unformatted" ] && say "c2: verdict gofmt PASS" || say "c2: verdict gofmt FAIL: $unformatted"

pins='^(TestWholeEnumIntoObjectOrEnumSlotLowers|TestNumberIntoObjectOrWholeEnumSlotLowers|TestNumberIntoWholeEnumSlotLowers|TestNumberIntoObjectOrMemberSlotStaysRefused|TestNumberIntoWholeEnumBesideAnotherMemberStaysRefused|TestSlotEnumPartIsItsNumericMembers)$'
go test -count=1 -timeout 30m -run "$pins" -v ./internal/lower/ > "$out/pins.log" 2>&1
code=$?
grep -E '^\s*--- (PASS|FAIL)' "$out/pins.log" | cut -c1-200 | head -20
[ "$code" = 0 ] && say "c2: verdict pins PASS" || { say "c2: verdict pins FAIL: exit $code"; grep -E '^\s+enum_slot_part_test.go' "$out/pins.log" | cut -c1-400 | head -10; }

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
		say "c2: verdict mutant $name BROKEN: the edit didn't apply"
		cp "$out/$name.orig" "$file"
		return
	fi
	go test -count=1 -timeout 30m -run "$tests" -v ./internal/lower/ > "$out/$name.log" 2>&1
	local code=$?
	cp "$out/$name.orig" "$file"
	grep -E '^\s*--- FAIL' "$out/$name.log" | cut -c1-200 | head -10
	if grep -q 'build failed\|\[setup failed\]\|cannot use\|undefined:\|declared and not used' "$out/$name.log"; then
		say "c2: verdict mutant $name BROKEN: it didn't compile"
		tail -n 20 "$out/$name.log"
		return
	fi
	local missing=""
	local IFS='|'
	for test in $must; do
		grep -qE "^\s*--- FAIL: $test " "$out/$name.log" || missing="$missing $test"
	done
	if [ "$code" != 0 ] && [ -z "$missing" ]; then
		say "c2: verdict mutant $name RED: $(grep -cE '^\s*--- FAIL' "$out/$name.log") failing"
	else
		say "c2: verdict mutant $name SURVIVED: exit $code, not failed:$missing"
	fi
}
mutant m1-object-counted internal/lower/enums.go \
	'	part := enumPart(flattenedMembers(l.checker.GetNonNullableType(slot)))' \
	'	part := flattenedMembers(l.checker.GetNonNullableType(slot))' \
	"$pins" \
	'TestWholeEnumIntoObjectOrEnumSlotLowers'
mutant m2-any-member-open internal/lower/enums.go \
	'	return !impure && len(whole) == 1 && whole[0].size >= 2 && whole[0].size == len(part)' \
	'	return !impure && (len(whole) > 0 || len(part) > 0) && l.numericEnum(l.enumIdentity(part[0]))' \
	"$pins" \
	'TestNumberIntoObjectOrMemberSlotStaysRefused/a_member|TestNumberIntoObjectOrMemberSlotStaysRefused/a_strict_subset|TestNumberIntoWholeEnumBesideAnotherMemberStaysRefused'
git -C "$root" status --short -- internal/lower | head -5

for package in lower ir; do
	go test -count=1 -timeout 40m ./internal/$package/ > "$out/whole-$package.log" 2>&1
	code=$?
	grep -E '^(--- FAIL|FAIL|ok|panic)' "$out/whole-$package.log" | cut -c1-300 | head -40
	[ "$code" = 0 ] && say "c2: verdict $package whole PASS" || say "c2: verdict $package whole FAIL: exit $code"
done

echo "c2: summary ($out)"
cat "$out/verdicts.log"
grep -q 'FAIL\|SURVIVED\|BROKEN' "$out/verdicts.log" && exit 1
exit 0
