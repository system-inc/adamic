source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
sandboxed_products ./internal/oracle ./internal/native > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
set -u
# #s03ndgw piece 1's proof: a class constructor a checked view admits (StaticsAdmitted) reads as Node reads it,
# and a write through such a view stops, after its value, on every road a write takes.
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; tail -60 "$TMPDIR/step.log"; ok=0; fi; }
want_red() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -ne 0 ]; then echo "PASS $name: red, as wanted"; grep -E "^\s+--- FAIL|error:|disagree|exit|StaticsAdmitted" "$TMPDIR/step.log" | head -16; else echo "FAIL $name: green, a survivor"; ok=0; fi; }
# A road no pin may take (which one a write takes is lowering's CheckedField): red is reported, green is a GAP.
road() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -ne 0 ]; then echo "PASS $name: red"; grep -E "^\s+--- FAIL" "$TMPDIR/step.log" | head -8; else echo "GAP $name: green, no pin takes this road"; fi; }
show() { grep -v "gate cache\|^=== " "$TMPDIR/step.log" | head -80; }
# mutate file anchor replacement: one fixed-string line edit, refused unless the anchor is there exactly once.
mutate() { python3 -I - "$1" "$2" "$3" << 'PYTHON'
import sys
path, anchor, replacement = sys.argv[1:]
text = open(path).read()
if text.count(anchor) != 1:
    sys.exit('mutant anchor appears %d times in %s' % (text.count(anchor), path))
open(path, 'w').write(text.replace(anchor, replacement))
PYTHON
}
# restore file: put head back from its saved copy and prove it.
restore() { cp "scratch-s03p1/$(basename "$(dirname "$1")")_$(basename "$1").head" "$1"; if cmp -s "scratch-s03p1/$(basename "$(dirname "$1")")_$(basename "$1").head" "$1"; then echo "PASS restored $1"; else echo "FAIL restore $1"; ok=0; fi; }
runtime=internal/native_runtime/c/view_unions_untagged.c
contracts=internal/lower/view_contracts.go
writes='TestViewStaticClassStops/(write|write_value_order|logical|compound|increment)\.ts'

# Native admission mutant 1: the C file as main had it, is_static refused in plain_matches and plain_slot. It has
# no adamic_view_static_read either, so the object-field pins go red at the link too.
cp scratch-s03p1/view_unions_untagged.c.main "$runtime"
grep -q statics_admitted "$runtime" && { echo "FAIL native mutant 1 not applied"; ok=0; }
want_red "native admission mutant 1 (view_unions_untagged.c at main)" go test -count=1 -run 'TestViewStaticClassAgreesWithNode|TestFX7StaticClassAgreesWithNode' ./internal/oracle/ -v
restore "$runtime"
# Native admission mutant 2: only the object branch's admission refused again, everything else of head kept.
mutate "$runtime" 'if (!contract->statics_admitted || plain_union_holds_callable(contracts, count, parent)) { goto complete; }' 'if (true || plain_union_holds_callable(contracts, count, parent)) { goto complete; }' || ok=0
want_red "native admission mutant 2 (plain_matches refuses every class constructor)" go test -count=1 -run 'TestViewStaticClassAgreesWithNode|TestFX7StaticClassAgreesWithNode' ./internal/oracle/ -v
restore "$runtime"

# The object-field road (ir.Program's ObjectReadContract): a class constructor at a field typed Inner itself.
agreeing='TestViewStaticClassAgreesWithNode/(tagged_count|string_static|optional_absent)\.ts'
# Native read mutant 1: adamic_view_static_read refuses every class constructor, admitted or not.
mutate "$runtime" '    if (admitted && owned == count) { return; }' '    if (false) { return; }' || ok=0
want_red "native read mutant 1 (adamic_view_static_read refuses every class constructor)" go test -count=1 -run "$agreeing" ./internal/oracle/ -v
restore "$runtime"
# Native read mutant 2: adamic_view_static_read admits every class constructor, owned statics or not.
mutate "$runtime" '    if (admitted && owned == count) { return; }' '    if (true) { return; }' || ok=0
want_red "native read mutant 2 (adamic_view_static_read admits every class constructor)" go test -count=1 -run 'TestViewStaticClassStops/inherited_field\.ts' ./internal/oracle/ -v
restore "$runtime"
# JavaScript read mutant 1: adamicViewField gets no statics, so it finds every class constructor a function.
file=internal/javascript/view_fields.go
mutate "$file" '	return ", [" + strings.Join(required, ", ") + "]"' '	return ""' || ok=0
want_red "javascript read mutant 1 (no statics for adamicViewField)" go test -count=1 -run "$agreeing" ./internal/oracle/ -v
restore "$file"
# JavaScript read mutant 2: adamicViewField admits every class constructor, owned statics or not.
file=internal/javascript/readiness.go
mutate "$file" 'adamicViewObject(value) || statics !== undefined && adamicViewStatics(value, statics)' 'adamicViewObject(value) || adamicViewConstructor(value)' || ok=0
want_red "javascript read mutant 2 (adamicViewField admits every class constructor)" go test -count=1 -run 'TestViewStaticClassStops/inherited_field\.ts' ./internal/oracle/ -v
restore "$file"

# Lower mutant: viewStaticsAdmitted reads an empty field list, so Named and Called turn true.
mutate "$contracts" 'for _, field := range contract.Fields {' 'for _, field := range []ir.ViewFieldContract{} {' || ok=0
want_red "lower mutant (an empty field list)" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v
restore "$contracts"

# Write guard mutants, each backend's guard dropped on every road at once: the write pins go red.
file=internal/native/view_unions_untagged.go
mutate "$file" '	if !e.program.StaticsAdmittedReceiver(statement.ViewReceiverTypeID) {' '	if true {' || ok=0
want_red "native write guard dropped on every road" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"
file=internal/javascript/view_fields.go
mutate "$file" '	if !e.program.StaticsAdmittedReceiver(statement.ViewReceiverTypeID) {' '	if true {' || ok=0
want_red "javascript write guard dropped on every road" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"
# Native's every SetProperty, checked or not, and its checked logical assignment, write through writeProperty.
file=internal/native/emit_statements.go
mutate "$file" '	e.staticViewWrite(statement, object)
' '' || ok=0
want_red "native writeProperty road" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"
# The roads lowering's CheckedField chooses between: each pin takes one of a pair.
file=internal/native/logical_assignment.go
mutate "$file" '		e.staticViewWrite(write, receiver)
' '' || ok=0
road "native unchecked logical road" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"
file=internal/javascript/view_fields.go
mutate "$file" '	return e.staticViewWrite(statement, object, value, write)' '	return write(object, value)' || ok=0
road "javascript checked write road (=, compound, increment, checked logical)" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"
file=internal/javascript/javascript.go
mutate "$file" 'e.staticViewWrite(statement, e.value(statement.Object), e.value(statement.Value), write)' 'write(e.value(statement.Object), e.value(statement.Value))' || ok=0
road "javascript unchecked write road" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"
file=internal/javascript/taste.go
mutate "$file" 'e.staticViewWrite(write, receiver, value, assign)' 'assign(receiver, value)' || ok=0
road "javascript unchecked logical road" go test -count=1 -run "$writes" ./internal/oracle/ -v
restore "$file"

[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
