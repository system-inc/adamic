set -u
# #0d56xpd rule (d) on main dacb2da2: a ?. read whose result the checker calls a scalar is held as the maybe it is
# (optional_scalar.go), a placeholder's ?. receiver is a presence test, and an unannotated const { f } = x is a
# placeholder slot. PART=a: gofmt, vet, the lower pins verbose, products, the oracle pins verbose, then mutants M1 to
# M7 (each one exact line replaced by python, checked replaced once, the pins run, put back from the head copy and
# checked put back). PART=d: lower whole. PART=b: oracle whole and native whole. PART=c: stage3/fixtures, ir,
# compilercensus and parallelcensus whole.
source /opt/adamic-block/environment.sh && echo "PASS environment.sh sourced" || { echo "FAIL environment.sh"; exit 1; }
source cloud/sandboxed-job.sh && echo "PASS sandboxed-job.sh sourced" || { echo "FAIL sandboxed-job.sh"; exit 1; }
for helper in sandboxed_npm sandboxed_products sandboxed_typescript_source; do
	declare -F "$helper" > /dev/null && echo "PASS $helper defined" || { echo "FAIL $helper undefined"; exit 1; }
done
work="${LOOM_JOB_SCRATCH:-${TMPDIR:-/tmp}}/proof0d56"
mkdir -p "$work" && echo "PASS scratch $work" || { echo "FAIL scratch $work"; exit 1; }
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$work/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$work/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$work/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$work/whole.log" | sort -u | head -40; grep -E "_test.go:[0-9]+:" "$work/whole.log" | head -20 | cut -c1-600; grep -qE "^(--- FAIL|panic:)" "$work/whole.log" || tail -25 "$work/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$work/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$work/step.log" && ok=0; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

lowerPins='TestOptionalScalarReadLowers'
oraclePins='TestOptionalScalarRead$|TestPlaceholderDestructuredBindingIsChecked'
if [ "$part" = a ]; then
	files="internal/lower/object.go internal/lower/optional_scalar.go internal/lower/placeholder.go internal/lower/switch_nullish_numeric.go"
	for f in $files; do cp "$f" "$work/$(basename $f).head"; done
	trap 'for f in $files; do cp "$work/$(basename $f).head" "$f"; done' EXIT
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/optional_scalar.go internal/lower/optional_scalar_test.go internal/lower/object.go internal/lower/placeholder.go internal/lower/placeholder_destructure.go internal/lower/collections.go internal/lower/assignments.go internal/lower/switch_nullish_numeric.go internal/oracle/optional_scalar_test.go internal/oracle/placeholder_destructure_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step "lower pins" go test -count=1 -run "$lowerPins" ./internal/lower/ -v
	subtests
	grep -E "_test.go:[0-9]+:" "$work/step.log" | head -20 | cut -c1-600
	step products sandboxed_products ./internal/oracle
	step "oracle pins" go test -count=1 -run "$oraclePins" ./internal/oracle/ -v
	subtests
	grep -E "_test.go:[0-9]+:" "$work/step.log" | head -30 | cut -c1-600

	# mutant name file old new: replace exactly one line, run both pins, report each failing subtest and its words, put back.
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
		{ go test -count=1 -run "$lowerPins" ./internal/lower/ -v; go test -count=1 -run "$oraclePins" ./internal/oracle/ -v; } > "$work/mutant.log" 2>&1
		reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$work/mutant.log" | tr '\n' ' ')
		if [ -z "$reds" ] && ! grep -qE "^FAIL|cannot|undefined:" "$work/mutant.log"; then echo "SURVIVED $name"; else echo "RED $name: $reds"; fi
		grep -E "^# |cannot|undefined:|_test.go:[0-9]+:" "$work/mutant.log" | head -6 | cut -c1-400
		cp "$work/$(basename $file).head" "$file"
		if cmp -s "$work/$(basename $file).head" "$file"; then echo "PASS restored $file after $name"; else echo "FAIL restore $file"; ok=0; fi
	}
	mutant M1 internal/lower/object.go \
		'		read := l.readObjectField(node, ir.Property{Object: object, Name: name, Of: of, Optional: true, Class: l.classOf(node)})' \
		'		read := l.readObjectField(node, ir.Property{Object: object, Name: name, Of: of, Optional: false, Class: l.classOf(node)})'
	mutant M2 internal/lower/optional_scalar.go \
		'	return ir.Coalesce{Value: read, Panic: ir.StringConstant{Index: l.constant(message)}, Of: of}' \
		'	return map[bool]ir.Expression{true: read}[message != ""]'
	mutant M3 internal/lower/placeholder.go \
		'		if l.placeholderChainObject(original) {' \
		'		if false && l.placeholderChainObject(original) {'
	mutant M4 internal/lower/placeholder.go \
		'	if access == nil || access.Kind != ast.KindPropertyAccessExpression || access.AsPropertyAccessExpression().Expression != at || access.AsPropertyAccessExpression().QuestionDotToken == nil {' \
		'	if access == nil || access.Kind != ast.KindPropertyAccessExpression || access.AsPropertyAccessExpression().Expression != at {'
	mutant M5 internal/lower/optional_scalar.go \
		'	if of == ir.Number && bitwiseOperand(node) {' \
		'	if false && of == ir.Number && bitwiseOperand(node) {'
	mutant M6 internal/lower/placeholder.go \
		'			if target.Kind == ast.KindObjectBindingPattern && node.AsVariableDeclaration().Type == nil && value != nil {' \
		'			if false && target.Kind == ast.KindObjectBindingPattern && node.AsVariableDeclaration().Type == nil && value != nil {'
	mutant M7 internal/lower/switch_nullish_numeric.go \
		'	numeric, nullish := false, storage == ir.MaybeNumber' \
		'	numeric, nullish := false, false'
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
