source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
sandboxed_products ./internal/oracle ./internal/native > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
set -u
# #24kvb6k's proof: a class constructor a checked view admits reads a static its base holds, as Node walks the
# constructor's prototype chain, and a required field named for Function's prototype, where a class declares it.
# A builtin (name, length, prototype) is found only on the class itself, an accessor anywhere on the chain is
# refused, and an optional field Function names keeps StaticsAdmitted false.
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; tail -60 "$TMPDIR/step.log"; ok=0; fi; }
want_red() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -ne 0 ]; then echo "PASS $name: red, as wanted"; grep -E "^\s+--- FAIL|error:|disagree|exit|StaticsAdmitted" "$TMPDIR/step.log" | head -16; else echo "FAIL $name: green, a survivor"; ok=0; fi; }
# A mutant no valid program reaches past lowering: red is reported, green is a GAP, never a failure.
road() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -ne 0 ]; then echo "PASS $name: red"; grep -E "^\s+--- FAIL" "$TMPDIR/step.log" | head -8; else echo "GAP $name: green, no pin reaches it"; fi; }
show() { grep -v "gate cache\|^=== " "$TMPDIR/step.log" | head -100; }
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
runtime=internal/native_runtime/c/view_unions_untagged.c
readiness=internal/javascript/readiness.go
contracts=internal/lower/view_contracts.go
# Each mutated file's pristine copy, taken from the tree this job runs, so restore puts back exactly what was proved.
pristine() { echo "scratch-k24/$(basename "$(dirname "$1")")_$(basename "$1").pristine"; }
for file in "$runtime" "$readiness" "$contracts"; do cp "$file" "$(pristine "$file")"; done
restore() { cp "$(pristine "$1")" "$1"; if cmp -s "$(pristine "$1")" "$1"; then echo "PASS restored $1"; else echo "FAIL restore $1"; ok=0; fi; }

# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
step gofmt bash -c 'difference=$(gofmt -d internal/ir/views.go internal/lower/view_contracts.go internal/lower/view_statics_admitted_test.go internal/native/view_unions_untagged.go internal/javascript/readiness.go internal/javascript/view_fields.go internal/javascript/view_unions_untagged.go internal/oracle/view_static_class_test.go); printf "%s\n" "$difference"; test -z "$difference"'
step vet go vet ./internal/ir/ ./internal/lower/ ./internal/native/ ./internal/javascript/ ./internal/oracle/ ./stage3/census/latent/statecopy/

# The pins at head, each program's verdict shown.
step "oracle pins" go test -count=1 -run 'TestViewStaticClass|TestFX7StaticClassAgreesWithNode' ./internal/oracle/ -v
show
step "lower pin" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v
show
step "review twin agrees" go test -count=1 -run 'TestReviewProgramsAgreeWithNode_[0-9]+/fxspptb_oct9_native_p67_static_class_union_member' ./internal/oracle/ -v
grep -E "^\s*--- (PASS|FAIL|SKIP)" "$TMPDIR/step.log" | head
grep -qE "^\s*--- PASS: TestReviewProgramsAgreeWithNode_[0-9]+/fxspptb_oct9_native_p67_static_class_union_member" "$TMPDIR/step.log" && echo "PASS review twin ran" || { echo "FAIL review twin did not run and pass"; ok=0; }
step "statecopy whole" go test -count=1 ./stage3/census/latent/statecopy/

# The views already pinned.
step "oracle views" go test -count=1 -run 'View|view|Interface|Fx7|fx7|FX7' ./internal/oracle/
step "native views" go test -count=1 -run 'View|view|FX7' ./internal/native/
step "javascript views" go test -count=1 -run 'View|view|FX7|Untagged|Logical|logical' ./internal/javascript/
step "lower views" go test -count=1 -run 'View|view|Untagged|Statics' ./internal/lower/

inherited='TestViewStaticClassAgreesWithNode/(inherited_field|three_level|inherited_optional|call_inherited)\.ts|TestViewStaticClassStops/inherited_count\.ts'

# Native mutant 1: the walk never follows static_parent, so a static only a base holds is refused, as piece 1 had it.
mutate "$runtime" '        if (builtin || held->class->static_parent == 0) { return static_refused; }' '        if (true || builtin || held->class->static_parent == 0) { return static_refused; }' || ok=0
want_red "native walk not following static_parent" go test -count=1 -run "$inherited" ./internal/oracle/ -v
restore "$runtime"

# Native mutant 2: the builtin rule dropped, so Sub's name walks to Base's B where Node reads Sub's own.
mutate "$runtime" '    bool builtin = static_builtin_name(field);' '    bool builtin = static_builtin_name(field) && false;' || ok=0
want_red "native builtin rule dropped" go test -count=1 -run 'TestViewStaticClassStops/name_inherited\.ts' ./internal/oracle/ -v
restore "$runtime"

# Native mutant 3: the accessor refusal dropped. Lowering stops a getter a checked contract names before any
# backend runs (static_getter, base_getter), so no pin is expected to reach it: it stays as the walk's own guard.
mutate "$runtime" '    if (adamic_accessor_find(object, field) != NULL) { return static_refused; }' '    if (false && adamic_accessor_find(object, field) != NULL) { return static_refused; }' || ok=0
road "native accessor refusal dropped" go test -count=1 -run 'TestViewStaticClass' ./internal/oracle/ -v
restore "$runtime"

# JavaScript mutant 1: the walk stops at the constructor itself, never its prototype.
mutate "$readiness" 'holder = Object.getPrototypeOf(holder)) {' 'holder = null) {' || ok=0
want_red "javascript walk not following the prototype" go test -count=1 -run "$inherited" ./internal/oracle/ -v
restore "$readiness"

# JavaScript mutant 2: the enumerable test dropped, so the emitted function's own name passes for a static.
mutate "$readiness" 'return Object.hasOwn(field, "value") && field.enumerable && !adamicFieldReadiness' 'return Object.hasOwn(field, "value") && !adamicFieldReadiness' || ok=0
want_red "javascript enumerable test dropped" go test -count=1 -run 'TestViewStaticClassStops/(name_field|name_inherited)\.ts' ./internal/oracle/ -v
restore "$readiness"

# Lower mutant: the optional test dropped, so no Function name keeps StaticsAdmitted false, and an undeclared
# optional call is absent natively where Node reads Function's.
mutate "$contracts" '|| field.Optional && l.checker.GetPropertyOfType(functionType, field.Name) != nil {' '|| false && l.checker.GetPropertyOfType(functionType, field.Name) != nil {' || ok=0
want_red "lower optional test dropped (oracle)" go test -count=1 -run 'TestViewStaticClassStops/call_optional_undeclared\.ts' ./internal/oracle/ -v
want_red "lower optional test dropped (lower pin)" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v
restore "$contracts"

[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
