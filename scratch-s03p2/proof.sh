source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
sandboxed_products ./internal/oracle ./internal/native > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
set -u
# #s03ndgw piece 2's proof: an array a checked view admits as an object (ArraysAdmitted) reads as Node reads
# it, each declared field absent, on the plain object-field road and the union-member road, and everything
# else about it keeps a stop: a name an array has, a required field, a write, Object.keys.
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; tail -60 "$TMPDIR/step.log"; ok=0; fi; }
want_red() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -ne 0 ]; then echo "PASS $name: red, as wanted"; grep -E "^\s+--- FAIL|error:|disagree|exit|ArraysAdmitted" "$TMPDIR/step.log" | head -16; else echo "FAIL $name: green, a survivor"; ok=0; fi; }
show() { grep -v "gate cache\|^=== " "$TMPDIR/step.log" | head -80; }
# mutate file anchor replacement: one fixed-string edit, refused unless the anchor is there exactly once.
mutate() { python3 -I - "$1" "$2" "$3" << 'PYTHON'
import sys
path, anchor, replacement = sys.argv[1:]
text = open(path).read()
if text.count(anchor) != 1:
    sys.exit('mutant anchor appears %d times in %s' % (text.count(anchor), path))
open(path, 'w').write(text.replace(anchor, replacement))
PYTHON
}
# pristine file: the saved copy of head's file, under scratch-s03p2/.
pristine() { echo "scratch-s03p2/$(basename "$(dirname "$1")")_$(basename "$1").head"; }
# restore file: put head back from its saved copy and prove it.
restore() { cp "$(pristine "$1")" "$1"; if cmp -s "$(pristine "$1")" "$1"; then echo "PASS restored $1"; else echo "FAIL restore $1"; ok=0; fi; }
runtime=internal/native_runtime/c/view_unions_untagged.c
object=internal/native_runtime/c/object.c
native_fields=internal/native/view_fields.go
native_writes=internal/native/view_unions_untagged.go
readiness=internal/javascript/readiness.go
javascript_fields=internal/javascript/view_fields.go
javascript_unions=internal/javascript/view_unions_untagged.go
contracts=internal/lower/view_contracts.go
# Every file a mutant touches is head's before any mutant runs, so a restore puts back what was proved.
for file in "$runtime" "$object" "$native_fields" "$native_writes" "$readiness" "$javascript_fields" "$javascript_unions" "$contracts"; do
	cmp -s "$(pristine "$file")" "$file" || { echo "FAIL $file differs from its pristine copy $(pristine "$file"): refresh the copy or restore the file first"; exit 1; }
done
agreeing='TestViewArrayObjectAgreesWithNode'
plain_agreeing='TestViewArrayObjectAgreesWithNode/(field_empty_interface|two_optional_fields)\.ts'
union_agreeing='TestViewArrayObjectAgreesWithNode/union_member\.ts'
plain_names='TestViewArrayObjectStops/(push_field|to_string_field|length_field)\.ts'
union_stops='TestViewArrayObjectStops/(index_field|required_field)\.ts'
writes='TestViewArrayObjectStops/widened_write\.ts'

# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
step gofmt bash -c 'difference=$(gofmt -d internal/ir/views.go internal/lower/view_contracts.go internal/lower/view_arrays_admitted_test.go internal/native/view_fields.go internal/native/view_unions_read.go internal/native/view_unions_untagged.go internal/javascript/javascript.go internal/javascript/readiness.go internal/javascript/taste.go internal/javascript/view_fields.go internal/javascript/view_unions_read.go internal/javascript/view_unions_untagged.go internal/oracle/view_array_object_test.go stage3/census/latent/statecopy/ir_fields.go); printf "%s\n" "$difference"; test -z "$difference"'
step vet go vet ./internal/ir/ ./internal/lower/ ./internal/native/ ./internal/javascript/ ./internal/oracle/ ./stage3/census/latent/statecopy/

# The pins at head: piece 2's, the array stops it shares a file with, and piece 1's, whose write guard and
# object read piece 2 extends.
step "new oracle pins" go test -count=1 -run 'TestViewArrayObject|TestViewArrayLiteralIsNotAnObject|TestViewTupleStillReadsAsATuple' ./internal/oracle/ -v
show
step "piece 1 pins" go test -count=1 -run 'TestViewStaticClass|TestFX7StaticClassAgreesWithNode' ./internal/oracle/ -v
grep -E "^\s*--- (PASS|FAIL|SKIP)" "$TMPDIR/step.log" | head -40
step "lower pins" go test -count=1 -run 'TestViewArraysAdmittedAsksArrayAndObject|TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v
show
step "statecopy whole" go test -count=1 ./stage3/census/latent/statecopy/
# ArraysAdmitted is set in finishViewStaticsAdmitted, piece 1's pass, so the scanning reference and the latent
# ledger are unchanged, and the cycle finder's equality compares both pipelines' IR, the new flag included.
step "pass tables" go test -count=1 -run 'TestScanningReferenceCallsEveryPass|TestCycleFinderEquality' ./internal/lower/
step "latent ledger" go test -count=1 -run 'TestLatentPassLedger' ./stage3/census/latent/

# The views already pinned, and the runtime readers piece 2 guards (keys, values, entries, hasOwn, freeze,
# assign, spread, JSON) by name.
step "oracle views" go test -count=1 -run 'View|view|Interface|Fx7|fx7|FX7' ./internal/oracle/
step "native views" go test -count=1 -run 'View|view|FX7|ErrorConstructionBytes' ./internal/native/
step "javascript views" go test -count=1 -run 'View|view|FX7|Untagged|Logical|logical' ./internal/javascript/
step "lower views" go test -count=1 -run 'View|view|Untagged' ./internal/lower/
step "oracle object readers" go test -count=1 -run 'Keys|Entries|HasOwn|Freeze|Assign|Spread|Stringify' ./internal/oracle/

# Union road, native. Mutant 1: plain_matches refuses every array its contract admits.
mutate "$runtime" 'result = array_lacks_fields(current->payload.reference, contract->fields, contract->field_count);' 'result = false;' || ok=0
want_red "native union mutant 1 (plain_matches refuses every array)" go test -count=1 -run "$union_agreeing" ./internal/oracle/ -v
restore "$runtime"
# Mutant 2: plain_matches admits every array for every object contract, ArraysAdmitted or not.
mutate "$runtime" 'if (contract->kind == 2 && contract->arrays_admitted && !contract->fixed_tuple' 'if (contract->kind == 2 && !contract->fixed_tuple' || ok=0
want_red "native union mutant 2 (plain_matches admits every array)" go test -count=1 -run "$union_stops" ./internal/oracle/ -v
restore "$runtime"

# Object-field road, native. Mutant 1: adamic_object_view_arrays refuses every array.
mutate "$object" '			if (lacking == count) { return *slot; }' '			if (false) { return *slot; }' || ok=0
want_red "native object mutant 1 (adamic_object_view_arrays refuses every array)" go test -count=1 -run "$plain_agreeing" ./internal/oracle/ -v
restore "$object"
# Mutant 2: every object read takes adamic_object_view_arrays, its contract admitting an array or not.
mutate "$native_fields" '	if !checked || !contract.ArraysAdmitted {' '	if !checked {' || ok=0
want_red "native object mutant 2 (every object read admits an array)" go test -count=1 -run "$plain_names" ./internal/oracle/ -v
restore "$native_fields"

# Object-field road, JavaScript. Mutant 1: adamicViewField refuses every array.
mutate "$readiness" '|| arrays !== undefined && adamicViewArray(value, arrays) :' ':' || ok=0
want_red "javascript object mutant 1 (adamicViewField refuses every array)" go test -count=1 -run "$plain_agreeing" ./internal/oracle/ -v
restore "$readiness"
# Mutant 2: adamicViewField gets arrays for every object read, its contract admitting one or not.
mutate "$javascript_fields" '	if !checked || !contract.ArraysAdmitted {' '	if !checked {' || ok=0
want_red "javascript object mutant 2 (every object read admits an array)" go test -count=1 -run "$plain_names" ./internal/oracle/ -v
restore "$javascript_fields"

# Union road, JavaScript. Mutant 1: matches refuses every array its contract admits.
mutate "$javascript_unions" 'return (contract.Fields || []).every(field=>!(field.Name in value));' 'return false;' || ok=0
want_red "javascript union mutant 1 (matches refuses every array)" go test -count=1 -run "$union_agreeing" ./internal/oracle/ -v
restore "$javascript_unions"
# Mutant 2: matches admits every array for every object contract.
mutate "$javascript_unions" 'contract.Kind===2 && contract.ArraysAdmitted===true && !contract.FixedTuple' 'contract.Kind===2 && !contract.FixedTuple' || ok=0
want_red "javascript union mutant 2 (matches admits every array)" go test -count=1 -run "$union_stops" ./internal/oracle/ -v
restore "$javascript_unions"

# Lower mutant 1: viewArraysAdmitted reads an empty field list, so Pushy, Lengthy, Demanding and the rest turn true.
mutate "$contracts" '	arrayType := l.checker.GetDeclaredTypeOfSymbol(array)
	for _, field := range contract.Fields {' '	arrayType := l.checker.GetDeclaredTypeOfSymbol(array)
	for _, field := range []ir.ViewFieldContract{} {' || ok=0
want_red "lower mutant 1 (an empty field list)" go test -count=1 -run 'TestViewArraysAdmittedAsksArrayAndObject' ./internal/lower/ -v
restore "$contracts"
# Lower mutant 2: the checker's names dropped, Array's and Object's (arrayType kept in use, so it compiles):
# push? and toString? turn admitted, and their stops go green, so the pins go red.
mutate "$contracts" ' || l.objectPrototypeName(field.Name) || l.checker.GetPropertyOfType(arrayType, field.Name) != nil {' ' || arrayType == nil {' || ok=0
want_red "lower mutant 2 (the name check dropped)" go test -count=1 -run 'TestViewArrayObjectStops/(push_field|to_string_field)\.ts' ./internal/oracle/ -v
restore "$contracts"

# Write guard mutants, each backend's dropped on every road at once: the widened write goes red.
mutate "$native_writes" '	if !e.program.ArraysAdmittedWrite(statement) {' '	if true {' || ok=0
want_red "native write guard dropped" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$native_writes"
mutate "$javascript_fields" '	if !e.program.ArraysAdmittedWrite(statement) {' '	if true {' || ok=0
want_red "javascript write guard dropped" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$javascript_fields"

[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
