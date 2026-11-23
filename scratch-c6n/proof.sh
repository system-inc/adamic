set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #snypn8c c6 on main 19c05dd6, with Lower's sharper rule: a member slot every visit reaches only past a field
# whose own contract isn't built (#rgrtt32) is walked past. PART=a: gofmt, vet, the lower pins verbose, mutants
# M1 to M5 (each one exact line replaced by python, checked replaced once, run, put back from the head copy and
# checked put back), the oracle's c6 pins, lower whole. PART=b: oracle whole and native whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

pins='TestViewCast'
if [ "$part" = a ]; then
	files="internal/lower/cast_overlap_view.go internal/lower/invariance.go"
	for f in $files; do cp "$f" "$TMPDIR/$(basename $f).head"; done
	trap 'for f in $files; do cp "$TMPDIR/$(basename $f).head" "$f"; done' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/cast_overlap_view.go internal/lower/view_cast_enum_slots_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/
	step "lower pins" go test -count=1 -run "$pins" ./internal/lower/ -v
	subtests
	# Every pin's error words, so a stop's wording is read, not assumed.
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | head -20

	# mutant name file old new: replace exactly one line, run the lower pins, report each failing subtest, put back.
	mutant() {
		name=$1 file=$2 old=$3 new=$4
		python3 -I - "$file" "$old" "$new" <<'PY'
import sys
path, old, new = sys.argv[1:4]
text = open(path).read()
if text.count(old) != 1:
    print(f"FAIL mutant: {old!r} occurs {text.count(old)} times"); sys.exit(1)
open(path, 'w').write(text.replace(old, new))
PY
		if [ $? -ne 0 ]; then ok=0; return; fi
		go test -count=1 -run "$pins" ./internal/lower/ -v > "$TMPDIR/mutant.log" 2>&1
		code=$?
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
		if [ $code -eq 0 ]; then echo "SURVIVED $name"; else echo "RED $name: $reds"; fi
		grep -E "^# |cannot|undefined:" "$TMPDIR/mutant.log" | head -5
		cp "$TMPDIR/$(basename $file).head" "$file"
		if cmp -s "$TMPDIR/$(basename $file).head" "$file"; then echo "PASS restored $file after $name"; else echo "FAIL restore $file"; ok=0; fi
	}
	mutant M1 internal/lower/cast_overlap_view.go \
		'	if found := l.widenedPast(source, target, slots, map[[2]*checker.Type]bool{}); found != nil {' \
		'	if found := l.widenedPast(source, target, nil, map[[2]*checker.Type]bool{}); found != nil {'
	mutant M2 internal/lower/invariance.go \
		'		if inside == nil || skip[viewed.Name] || past[viewCastSlot{to, viewed.Name}] {' \
		'		if inside == nil || skip[viewed.Name] || past != nil {'
	mutant M3 internal/lower/cast_overlap_view.go \
		'		through := path.readonly && l.checker.IsReadonlySymbol(viewed)' \
		'		through := path.readonly'
	mutant M4 internal/lower/cast_overlap_view.go \
		'			if field.contract != 0 && strings.HasPrefix(l.result.ViewContracts[field.contract-1].Unsupported, unbuiltFieldFamily) {' \
		'			if true || strings.HasPrefix(l.result.ViewContracts[field.contract-1].Unsupported, unbuiltFieldFamily) {'
	mutant M5 internal/lower/cast_overlap_view.go \
		'				field.behind = true' \
		'				field.behind = path.behind'

	step products sandboxed_products ./internal/oracle
	step "oracle pins" go test -count=1 -run 'TestViewCastEnumSlots' ./internal/oracle/ -v
	subtests
	whole "lower whole" go test -count=1 -timeout 40m ./internal/lower/
fi
if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle
	whole "oracle whole" go test -count=1 -timeout 40m ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi
if [ "$part" = c ]; then
	whole "stage3/fixtures whole" go test -count=1 -timeout 40m ./stage3/fixtures/
	whole "ir whole" go test -count=1 -timeout 40m ./internal/ir/
	whole "compilercensus whole" go test -count=1 -timeout 40m ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 -timeout 40m ./cmd/adamic-gate/parallelcensus/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
