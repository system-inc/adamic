set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #24kvb6k re-cut onto the Now (92554beb): the lift (a1e16123) walks a class constructor's statics up its chain, and
# lowering keeps StaticsAdmitted false for an optional field named like any static method the program declares, which
# the run time keeps with no name. PART=a: gofmt, vet, the pins at head, then each mutant from a whole-file copy in
# scratch-k24r/, put back from the head copy and checked put back. PART=b: the views already pinned, then lower,
# oracle, native (ADAMIC_TEST_WASI=1), javascript, compilercensus and parallelcensus whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; ok=0; fi; }
# subtests prints each verbose subtest's verdict as its own line, PASS or FAIL: a skipped one proved nothing.
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
# red runs a test that must fail under a mutant, then asks its log for each fixed string given after --.
red() {
	name=$1; shift
	command=()
	while [ "$1" != "--" ]; do command+=("$1"); shift; done
	shift
	"${command[@]}" > "$TMPDIR/mutant.log" 2>&1
	if [ $? -eq 0 ]; then echo "FAIL $name: green, a survivor"; ok=0; return; fi
	for want in "$@"; do
		if grep -qF -- "$want" "$TMPDIR/mutant.log"; then echo "PASS $name: red, saying $want"; else echo "FAIL $name: red, but never saying $want"; grep -E "^\s+--- FAIL|_test.go:[0-9]+:" "$TMPDIR/mutant.log" | head -12; ok=0; fi
	done
}
# road runs a mutant no valid program reaches past lowering: red or green, it's reported, never a failure.
road() { name=$1; shift; "$@" > "$TMPDIR/mutant.log" 2>&1; if [ $? -ne 0 ]; then echo "PASS $name: red"; else echo "PASS $name: green, no pin reaches it, as a1e16123's proof found"; fi; }

contracts=internal/lower/view_contracts.go
runtime=internal/native_runtime/c/view_unions_untagged.c
readiness=internal/javascript/readiness.go
head_copy() { echo "$TMPDIR/head_$(basename "$1")"; }
# put_back restores a mutated file from its head copy and checks it's byte for byte the head again.
put_back() { cp "$(head_copy "$1")" "$1"; if cmp -s "$(head_copy "$1")" "$1"; then echo "PASS restored $1"; else echo "FAIL restore $1"; ok=0; fi; }
# mutate puts a whole-file mutant in place, refused unless it differs from head in exactly one line, so a mutant cut
# against an older head never carries that head's other lines back in.
mutate() {
	target=$1 mutant=scratch-k24r/$2
	changed=$(diff "$(head_copy "$target")" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $2 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "$target"
	if cmp -s "$mutant" "$target"; then echo "PASS mutant $2 in place"; else echo "FAIL mutant $2 not in place"; ok=0; return 1; fi
}

if [ "$part" = a ]; then
	for file in "$contracts" "$runtime" "$readiness"; do cp "$file" "$(head_copy "$file")"; done
	# A killed run still leaves the tree as proved.
	trap 'for file in "$contracts" "$runtime" "$readiness"; do cp "$(head_copy "$file")" "$file"; done' EXIT

	# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
	step gofmt bash -c 'difference=$(gofmt -d internal/ir/views.go internal/lower/class_static.go internal/lower/lower.go internal/lower/view_contracts.go internal/lower/view_statics_admitted_test.go internal/native/view_unions_untagged.go internal/javascript/readiness.go internal/javascript/view_fields.go internal/javascript/view_unions_untagged.go internal/oracle/view_static_class_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/ir/ ./internal/lower/ ./internal/oracle/ ./internal/native/ ./internal/javascript/
	step products sandboxed_products ./internal/oracle

	# The pins at head, each program's verdict on its own line.
	step "view_static_class pins" go test -count=1 -run 'TestViewStaticClass' ./internal/oracle/ -v
	subtests
	step "lower pin" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v
	subtests

	# M1: the static method rule dropped from viewStaticsAdmitted. Each program's contract admits Klass, the walk finds
	# no slot named extra, and native reads undefined where Node reads Klass's method (optional_method) or undefined
	# itself (optional_method_unrelated, a stop on purpose); the lower pin finds Maybe admitted beside Tool.
	if mutate "$contracts" view_contracts.go.static_method.mutant; then
		red "M1 static method rule dropped, optional_method" go test -count=1 -run 'TestViewStaticClassStops/^optional_method\.ts$' ./internal/oracle/ -v -- 'native: ' 'stdout "before\nundefined\n"'
		red "M1 static method rule dropped, optional_method_unrelated" go test -count=1 -run 'TestViewStaticClassStops/^optional_method_unrelated\.ts$' ./internal/oracle/ -v -- 'native: ' 'stdout "before\nundefined\n"'
		red "M1 static method rule dropped, lower pin" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v -- "Tool's static extra: Maybe: StaticsAdmitted true, want false"
	fi
	put_back "$contracts"

	# M2: native's walk without its accessor refusal, as a1e16123's proof had it. Lowering stops a getter a checked
	# contract names before any backend runs (static_getter, base_getter, optional_getter), so no pin is expected to
	# reach it: it stays as the walk's own guard.
	if mutate "$runtime" view_unions_untagged.c.accessor.mutant; then
		road "M2 native accessor refusal dropped" go test -count=1 -run 'TestViewStaticClass' ./internal/oracle/ -v
	fi
	put_back "$runtime"

	# a1e16123's other mutants, carried as they were.
	inherited='TestViewStaticClassAgreesWithNode/^(inherited_field|three_level|inherited_optional|optional_inherited|call_inherited)\.ts$|TestViewStaticClassStops/^inherited_count\.ts$'
	# The native walk never follows static_parent, so a static only a base holds is refused, as piece 1 had it.
	if mutate "$runtime" view_unions_untagged.c.parent.mutant; then
		red "native walk not following static_parent" go test -count=1 -run "$inherited" ./internal/oracle/ -v -- '--- FAIL: TestViewStaticClassAgreesWithNode/optional_inherited.ts' '--- FAIL: TestViewStaticClassAgreesWithNode/three_level.ts'
	fi
	put_back "$runtime"
	# The builtin rule dropped, so Sub's name walks to Base's B where Node reads Sub's own.
	if mutate "$runtime" view_unions_untagged.c.builtin.mutant; then
		red "native builtin rule dropped" go test -count=1 -run 'TestViewStaticClassStops/^name_inherited\.ts$' ./internal/oracle/ -v -- '--- FAIL: TestViewStaticClassStops/name_inherited.ts'
	fi
	put_back "$runtime"
	# The JavaScript walk stops at the constructor itself, never its prototype.
	if mutate "$readiness" readiness.go.prototype.mutant; then
		red "javascript walk not following the prototype" go test -count=1 -run "$inherited" ./internal/oracle/ -v -- 'javascript: ' '--- FAIL: TestViewStaticClassAgreesWithNode/optional_inherited.ts'
	fi
	put_back "$readiness"
	# The JavaScript enumerable test dropped, so the emitted function's own name passes for a static.
	if mutate "$readiness" readiness.go.enumerable.mutant; then
		red "javascript enumerable test dropped" go test -count=1 -run 'TestViewStaticClassStops/^(name_field|name_inherited)\.ts$' ./internal/oracle/ -v -- 'javascript: '
	fi
	put_back "$readiness"
	# Lowering's Function test for an optional field dropped, the static method rule kept: an undeclared optional call
	# is absent natively where Node reads Function's, and the lower pin finds MaybeCalled admitted.
	if mutate "$contracts" view_contracts.go.optional.mutant; then
		red "lower optional Function test dropped (oracle)" go test -count=1 -run 'TestViewStaticClassStops/^call_optional_undeclared\.ts$' ./internal/oracle/ -v -- 'native: '
		red "lower optional Function test dropped (lower pin)" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v -- 'MaybeCalled: StaticsAdmitted true, want false'
	fi
	put_back "$contracts"
	trap - EXIT
fi

if [ "$part" = b ]; then
	step npm sandboxed_npm stage3/api stage3/typescript
	step products sandboxed_products ./internal/lower ./internal/oracle ./internal/native ./internal/javascript ./internal/compilercensus
	# The views already pinned, as a1e16123's proof ran them.
	step "oracle views" go test -count=1 -run 'View|view|Interface|Fx7|fx7|FX7' ./internal/oracle/
	step "native views" go test -count=1 -run 'View|view|FX7' ./internal/native/
	step "javascript views" go test -count=1 -run 'View|view|FX7|Untagged|Logical|logical' ./internal/javascript/
	whole "lower whole" go test -count=1 ./internal/lower/
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 ./internal/native/
	whole "javascript whole" go test -count=1 ./internal/javascript/
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi

if [ "$part" != a ] && [ "$part" != b ]; then echo "FAIL PART=$part names no part, want a or b"; ok=0; fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
