set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #hae423v waves 1 and 2: a closed range type ({ pos; end }) is held as a Union in a program that declares an
# interface with the NodeArray layout (range_types.go), every IR expression taking an object operand is classed for
# it (range_unions.go), and a NodeArray going where one is held is served, pos and end read, written, written when
# optional and destructured through it by a kind dispatch (range_fields.go).
# PART=a: gofmt, vet, the lower and oracle pins verbose with each subtest's verdict, the four mutants red from
# whole-file copies in scratch-ha1/, each put back and checked put back, then lower whole and the latent census
# package whole, whose pass ledger and statecopy guard mirror lowering's tail and state. PART=b: oracle whole and
# native whole (ADAMIC_TEST_WASI=1), so each part fits one job.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
# subtests prints each verbose test's and subtest's verdict as its own line, PASS or FAIL: a skipped one proved nothing.
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

targets="range_types.go range_unions.go range_fields.go"
heads="$TMPDIR/heads"
# put_back restores a mutated file from its head copy and checks it's byte for byte the head again.
put_back() { cp "$heads/$1" "internal/lower/$1"; if cmp -s "$heads/$1" "internal/lower/$1"; then echo "PASS restored $1"; else echo "FAIL restore $1"; ok=0; fi; }
# mutate puts a whole-file mutant of target in place, refused unless it differs from head in exactly one line, so a
# mutant cut against an older head never carries that head's other lines back in.
mutate() {
	target=$1 mutant=scratch-ha1/$2
	changed=$(diff "$heads/$target" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $2 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "internal/lower/$target"
	if cmp -s "$mutant" "internal/lower/$target"; then echo "PASS mutant $2 in place"; else echo "FAIL mutant $2 not in place"; ok=0; return 1; fi
}

echo "box: $(hostname) at $(git rev-parse HEAD) PART=$part"
# The whole runs read stage3's TypeScript (TestAssignmentProof*), which the box lays only on asking.
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = a ]; then
	mkdir -p "$heads"
	for file in $targets; do cp "internal/lower/$file" "$heads/$file"; done
	# A killed run still leaves the tree as proved.
	trap 'for file in $targets; do cp "$heads/$file" "internal/lower/$file"; done' EXIT

	# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/range_types.go internal/lower/range_unions.go internal/lower/range_fields.go internal/lower/range_unions_test.go internal/lower/expression.go internal/lower/lower.go internal/lower/object.go internal/lower/class.go internal/lower/class_accessors.go internal/lower/collections.go internal/lower/shape_fit.go internal/lower/cycles_scanning_reference_test.go internal/oracle/range_unions_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/lower ./internal/oracle ./stage3/census/latent

	# The pins at head, each verdict on its own line. Lower: the table over internal/ir's source, the
	# representation with and without a NodeArray, the read's dispatch, the pass's narrow stop, the scanning
	# reference still mirroring the pipeline, and the views' own source walk. Oracle: the eight programs held to
	# Node natively, sanitized and on JavaScript (equality and an object literal among them), and the two roads
	# that still stop.
	step "lower pins" go test -count=1 -run '^(TestEveryOperandIsClassedForRangeUnions|TestRangeTypeHeldAsUnionWithNodeArrayLayout|TestRangeUnionFieldReadDispatches|TestRangeUnionNarrowStops|TestScanningReferenceCallsEveryPass|TestEveryOperandIsClassedForViews)$' ./internal/lower/ -v
	subtests
	step "oracle pins" go test -count=1 -run '^(TestRangeUnionsAgreeWithNode|TestRangeUnionsStopWhereNotServed)$' ./internal/oracle/ -v
	subtests

	# M1 (wave 1): the bit's walk reads only each file's top-level statements, so a NodeArray declared in a
	# function's body is missed, and that program's TextRange is held as an object.
	if mutate range_types.go range_types.go.walk_misses.mutant; then
		red "M1 walk misses" go test -count=1 -run '^TestRangeTypeHeldAsUnionWithNodeArrayLayout$' ./internal/lower/ -v -- "--- FAIL: TestRangeTypeHeldAsUnionWithNodeArrayLayout/a_NodeArray_in_a_function's_body" 'TextRange is held as'
	fi
	put_back range_types.go

	# M2 (wave 1): TypeOf taken out of the table's dispatches, so the source walk finds it classed by nobody.
	if mutate range_unions.go range_unions.go.class_removed.mutant; then
		red "M2 class removed" go test -count=1 -run '^TestEveryOperandIsClassedForRangeUnions$' ./internal/lower/ -v -- '--- FAIL: TestEveryOperandIsClassedForRangeUnions' 'rangeUnionOf(ir.TypeOf'
	fi
	put_back range_unions.go

	# M3 (wave 2, Lower's Must 3): the read's object arm built as a dynamic read, a node finishAccessors never
	# turns into its accessor dispatcher, so the class's get pos() isn't run through the TextRange.
	if mutate range_fields.go range_fields.go.arm_unseen.mutant; then
		red "M3 object arm unseen" go test -count=1 -run '^TestRangeUnionsAgreeWithNode$/^getter_through_range\.a$' ./internal/oracle/ -v -- '--- FAIL: TestRangeUnionsAgreeWithNode/getter_through_range.a'
	fi
	put_back range_fields.go

	# M4 (wave 2): the read's array arm narrowed unchecked, which checkRangeUnionReads stops at lowering as an
	# unchecked narrow of a union that may be a range type's value.
	if mutate range_fields.go range_fields.go.array_unchecked.mutant; then
		red "M4 array arm unchecked" go test -count=1 -run '^TestRangeUnionsAgreeWithNode$/^nodearray_through_range\.a$' ./internal/oracle/ -v -- '--- FAIL: TestRangeUnionsAgreeWithNode/nodearray_through_range.a' 'a narrow of a union taken unasked'
	fi
	# M5 (Lower's condition): the in test classed as stopping again, so a NodeArray's in test, whose boxed operand
	# counts as a range value, stops at lowering.
	if mutate range_unions.go range_unions.go.in_stops.mutant; then
		red "M5 in test stops" go test -count=1 -run '^TestRangeUnionInTestOfANodeArrayLowers$' ./internal/lower/ -v -- '--- FAIL: TestRangeUnionInTestOfANodeArrayLowers' 'an in test'
		put_back range_unions.go
	fi
	put_back range_fields.go
	trap - EXIT

	whole "lower whole" go test -count=1 ./internal/lower/
	# The latent census mirrors lowering's tail (its pass ledger) and copies its state (statecopy's guard), so a
	# pass or a field either wave adds without its row reds here.
	whole "latent census whole" go test -count=1 ./stage3/census/latent/...
fi

if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle ./internal/native
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi

if [ "$part" = c ]; then
	sandboxed_typescript_source
	whole "stage3/fixtures whole" go test -count=1 ./stage3/fixtures/
	whole "ir whole" go test -count=1 ./internal/ir/
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi

if [ "$part" != a ] && [ "$part" != b ] && [ "$part" != c ]; then echo "FAIL PART=$part names no part, want a, b or c"; ok=0; fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
