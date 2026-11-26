set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #snypn8c c1 on main c35c1023: an && result its consumer drops when falsy (the left of an ||, a !, a condition) is
# never judged against a slot (andFalsyDropped), and a left's truthy-only literal arms never reach the result
# (falsyArms). PART=a: gofmt, vet, the lower pins verbose, mutants M1 to M6 (each one exact line replaced by python,
# checked replaced once, run, put back from the head copy and checked put back), the string-left probe, the oracle's
# c1 pins. PART=d: lower whole. PART=b: oracle whole and native whole. PART=c: stage3/fixtures, ir, compilercensus
# and parallelcensus whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -E "_test.go:[0-9]+:" "$TMPDIR/whole.log" | head -20; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

pins='TestAndFalsyDropped|TestAndLeft'
if [ "$part" = a ]; then
	files="internal/lower/logical_and_left.go internal/lower/invariance.go internal/lower/and_falsy_dropped_test.go"
	for f in $files; do cp "$f" "$TMPDIR/$(basename $f).head"; done
	trap 'for f in $files; do cp "$TMPDIR/$(basename $f).head" "$f"; done' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/logical_and_left.go internal/lower/logical_and_left_test.go internal/lower/and_falsy_dropped_test.go internal/lower/invariance.go internal/oracle/and_falsy_dropped_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step "lower pins" go test -count=1 -run "$pins" ./internal/lower/ -v
	subtests
	# Every pin's error words, so a stop's wording is read, not assumed.
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | head -30 | cut -c1-600

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
		grep -E "^# |cannot|undefined:|declared and not used|imported and not used" "$TMPDIR/mutant.log" | head -5
		grep -E "_test.go:[0-9]+:" "$TMPDIR/mutant.log" | head -8 | cut -c1-400
		cp "$TMPDIR/$(basename $file).head" "$file"
		if cmp -s "$TMPDIR/$(basename $file).head" "$file"; then echo "PASS restored $file after $name"; else echo "FAIL restore $file"; ok=0; fi
	}
	mutant M1 internal/lower/logical_and_left.go \
		'	if parent == nil {' \
		'	if parent == nil || true {'
	mutant M1a internal/lower/invariance.go \
		'			if andFalsyDropped(and) {' \
		'			if false {'
	mutant M1b internal/lower/invariance.go \
		'			if !andFalsyDropped(node) {' \
		'			if true {'
	mutant M2 internal/lower/logical_and_left.go \
		'		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken && parent.AsBinaryExpression().Left == child' \
		'		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken'
	mutant M3 internal/lower/logical_and_left.go \
		'		return parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken && parent.AsBinaryExpression().Left == child' \
		'		return (parent.AsBinaryExpression().OperatorToken.Kind == ast.KindBarBarToken || parent.AsBinaryExpression().OperatorToken.Kind == ast.KindQuestionQuestionToken) && parent.AsBinaryExpression().Left == child'
	mutant M4 internal/lower/logical_and_left.go \
		'		if l.truthyLiteral(arm) && !l.openEnumMember(arm, whole, impure) {' \
		'		if false && !l.openEnumMember(arm, whole, impure) {'
	mutant M5 internal/lower/and_falsy_dropped_test.go \
		'	walks := []string{"freshOrWidened", "refuseWidening"}' \
		'	walks := []string{"refuseWidening"}'
	mutant M6 internal/lower/logical_and_left.go \
		'	return impure || slices.ContainsFunc(whole, func(enum wholeNumericEnum) bool { return enum.symbol == l.enumIdentity(arm) })' \
		'	return impure && slices.ContainsFunc(whole, func(enum wholeNumericEnum) bool { return enum.symbol == l.enumIdentity(arm) })'
	# Part 2's string row, probed: does tsc accept a string literal on an && left, and what does Adamic say?
	go run ./cmd/adamic types scratch-c1/string_left.a > "$TMPDIR/string.log" 2>&1; echo "string_left.a types: exit $?"; grep -v "^\s*$" "$TMPDIR/string.log" | head -12 | cut -c1-400
	go run ./cmd/adamic c scratch-c1/string_left.a > /dev/null 2> "$TMPDIR/string-c.log"; echo "string_left.a c: exit $?"; head -c 1200 "$TMPDIR/string-c.log"; echo

	step products sandboxed_products ./internal/oracle
	step "oracle pins" go test -count=1 -run 'TestAndFalsyDroppedAgreesWithNode|TestAndLeft' ./internal/oracle/ -v
	subtests
	grep -E "_test.go:[0-9]+:" "$TMPDIR/step.log" | head -20 | cut -c1-600
fi
if [ "$part" = d ]; then
	# The census products read the TypeScript source, which must be the job's own copy (sandboxed_typescript_source).
	step "typescript source" sandboxed_typescript_source
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
