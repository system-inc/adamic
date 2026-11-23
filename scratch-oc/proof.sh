set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #cdpxjmb on emit/optional-chain-guard-1 (2f492e0c on main ac44317c), with Lower's absent-field row in fields.a: an optional chain's work after its nullish receiver runs under one guard
# (internal/lower/optional_chain.go, docs/rulings.md row 207), calls and nested fields alike, and m?.size on a
# Map or a Set is number | undefined (#8zcm7g4), with Lower's three musts: a step strips only the undefined the
# chain adds, every raw ?. reader is listed (TestOptionalChainReadersGoThroughTheGuard), and parentheses end a
# chain. PART=a: gofmt, vet, the pins at head with each subtest's verdict (the oracle's testdata/optional_chain
# held to Node on every backend, its asserted/ programs held to the assertion's words, the lower pins, and the
# pins the change flips: conformance's member_call.a and the stage1 gaps), the nine mutants, each from a
# whole-file copy in scratch-oc/, put back from the head copy and checked put back, then lower, ir,
# compilercensus and parallelcensus whole. PART=b: oracle whole and native whole (ADAMIC_TEST_WASI=1), so each part fits one job.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
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
		if grep -qF -- "$want" "$TMPDIR/mutant.log"; then echo "PASS $name: red, saying $want"; else echo "FAIL $name: red, but never saying $want"; grep -E "^\s+--- FAIL|_test.go:[0-9]+:|^panic:" "$TMPDIR/mutant.log" | head -12; ok=0; fi
	done
}

# Each mutated file has a head copy, taken before any mutant goes in.
chain=internal/lower/optional_chain.go
backend=internal/javascript/javascript.go
presence=internal/lower/eep_presence.go
chain_head="$TMPDIR/head_optional_chain.go"
backend_head="$TMPDIR/head_javascript.go"
presence_head="$TMPDIR/head_eep_presence.go"
# put_back restores a mutated file from its head copy and checks it's byte for byte the head again.
put_back() { target=$1; head_copy=$2; cp "$head_copy" "$target"; if cmp -s "$head_copy" "$target"; then echo "PASS restored $target"; else echo "FAIL restore $target"; ok=0; fi; }
# mutate puts a whole-file mutant in place over target, refused unless it differs from target's head copy in
# exactly one line, so a mutant cut against an older head never carries that head's other lines back in.
mutate() {
	target=$1; head_copy=$2; mutant=scratch-oc/$3
	changed=$(diff "$head_copy" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $3 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "$target"
	if cmp -s "$mutant" "$target"; then echo "PASS mutant $3 in place"; else echo "FAIL mutant $3 not in place"; ok=0; return 1; fi
}
oracle_pin=(go test -count=1 -run '^TestOptionalChainsAgreeWithNode$' ./internal/oracle/ -v)

echo "box: $(hostname) at $(git rev-parse HEAD) PART=$part"
# The whole runs read stage3's TypeScript (TestAssignmentProof*), which the box lays only on asking.
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = a ]; then
	cp "$chain" "$chain_head"
	cp "$backend" "$backend_head"
	cp "$presence" "$presence_head"
	# A killed run still leaves the tree as proved.
	trap 'cp "$chain_head" "$chain"; cp "$backend_head" "$backend"; cp "$presence_head" "$presence"' EXIT

	# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/optional_chain.go internal/lower/optional_chain_test.go internal/lower/optional_chain_readers_test.go internal/lower/slot_type_test.go internal/lower/eep_presence.go internal/lower/element_access_tuples.go internal/lower/enums.go internal/lower/merged_fields.go internal/lower/phantom_brands.go internal/lower/optional_indexing_map.go internal/lower/object.go internal/lower/expression.go internal/lower/class.go internal/lower/statements.go internal/lower/typed_arrays.go internal/lower/optional_intrinsic.go internal/lower/optional_join.go internal/lower/optional_method_call.go internal/lower/lower.go internal/lower/optional_value_call_test.go internal/lower/optional_indexing_test.go internal/lower/optional_intrinsic_test.go internal/oracle/optional_chain_test.go stage1/cohere/markdownblocks/gaps_test.go stage1/cohere/static_single_assignment/gaps_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/ ./internal/conformance/ ./stage1/cohere/markdownblocks/ ./stage1/cohere/static_single_assignment/
	step products sandboxed_products ./internal/oracle ./internal/lower

	# The pins at head, each program's verdict on its own line.
	step "oracle pins" go test -count=1 -run '^(TestOptionalChainsAgreeWithNode|TestOptionalChainAssertionsKeepTheirChecks|TestOptionalMethodCallsAgreeWithNode|TestOptionalMethodCallThroughAViewIsNotYet|TestMethodBindRefusals)$' ./internal/oracle/ -v
	subtests
	step "lower pins" go test -count=1 -run '^(TestOptionalChainStepsHaveTheirOwnTypes|TestOptionalChainGapsStop|TestOptionalChainReadersGoThroughTheGuard|TestReferenceWalksResolveShorthandNames|TestRefusalRulesReadTheirSlotThroughSlotType|TestOptionalValueCallGapsStop|TestOptionalValueCallsAgree|TestOptionalIndexingKeepsUnsupportedStorageNotYet|TestOptionalIntrinsicArgumentBoundariesRefuse|TestOptionalIntrinsicKeepsMaybeScalarRefusals|TestOptionalMethodCallGapsStop|TestOptionalMethodCallsLower|TestNodeFSDirentOptionalFunctionsAndContinuationsStop)$' ./internal/lower/ -v
	subtests
	step "conformance pins" go test -count=1 -run '^TestEveryCaseKeepsTheSpecsPromise$' ./internal/conformance/ -v
	subtests
	step "stage1 gap pins" go test -count=1 -run '^(TestParserRepresentationProbes|TestEachGapStandsWhereGapsMdSaysItDoes)$' ./stage1/cohere/markdownblocks/ ./stage1/cohere/static_single_assignment/ -v
	subtests

	# M1: the steps run on the missing receiver too, so an undefined one reads a field of nothing.
	if mutate "$chain" "$chain_head" optional_chain.go.m1_steps_unguarded.mutant; then
		red "M1 steps unguarded" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/chain.a'
	fi
	put_back "$chain" "$chain_head"
	# M2: the missing chain's value is the receiver itself, so a null receiver prints null.
	if mutate "$chain" "$chain_head" optional_chain.go.m2_receiver_as_value.mutant; then
		red "M2 receiver as the value" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/chain.a'
	fi
	put_back "$chain" "$chain_head"
	# M3: the steps take the receiver's expression, not the held read, so it's evaluated again and reads counts it.
	if mutate "$chain" "$chain_head" optional_chain.go.m3_receiver_again.mutant; then
		red "M3 receiver evaluated again" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/chain.a'
	fi
	put_back "$chain" "$chain_head"
	# M4: the steps run before the guard, arguments included.
	if mutate "$chain" "$chain_head" optional_chain.go.m4_steps_before_guard.mutant; then
		red "M4 steps before the guard" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/chain.a'
	fi
	put_back "$chain" "$chain_head"
	# M5: the JavaScript backend looks a method up after its arguments run, so callTalk calls the replacement.
	if mutate "$backend" "$backend_head" javascript.go.m5_callee_after_arguments.mutant; then
		red "M5 callee after the arguments" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/chain.a' 'javascript: '
	fi
	put_back "$backend" "$backend_head"
	# M6: only undefined is tested, so a null receiver's steps run.
	if mutate "$chain" "$chain_head" optional_chain.go.m6_undefined_only.mutant; then
		red "M6 undefined only" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/chain.a'
	fi
	put_back "$chain" "$chain_head"
	# M7: a missing collection's size is 0, not undefined.
	if mutate "$chain" "$chain_head" optional_chain.go.m7_size_zero.mutant; then
		red "M7 size zero" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/size.a'
	fi
	put_back "$chain" "$chain_head"
	# M8: every step stripped of undefined, a field's own declared one included, so fields.a's n?: number reads
	# as a number natively.
	if mutate "$chain" "$chain_head" optional_chain.go.m8_every_step_stripped.mutant; then
		red "M8 every step stripped" "${oracle_pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/fields.a'
	fi
	put_back "$chain" "$chain_head"
	# M9: one reader put back to the raw ?., which the readers' source test names.
	if mutate "$presence" "$presence_head" eep_presence.go.m9_raw_reader.mutant; then
		red "M9 a raw reader" go test -count=1 -run '^TestOptionalChainReadersGoThroughTheGuard$' ./internal/lower/ -v -- '--- FAIL: TestOptionalChainReadersGoThroughTheGuard' 'in eepPropertyPresence'
	fi
	put_back "$presence" "$presence_head"
	trap - EXIT

	whole "lower whole" go test -count=1 ./internal/lower/
	whole "ir whole" go test -count=1 ./internal/ir/
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi

if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle ./internal/native
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi

if [ "$part" != a ] && [ "$part" != b ]; then echo "FAIL PART=$part names no part, want a or b"; ok=0; fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
