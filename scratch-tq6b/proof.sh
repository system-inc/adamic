set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #tq6yzqy: an intersection field held as an optional boolean by the plain rule, as every member that has it holds
# it, is admitted (objectIntersection, intersectionMembersHoldMaybeBoolean). gofmt, vet, the pins at head with each
# subtest's verdict, the Record trace's probes on Node and both backends, then each mutant from a whole-file copy in
# scratch-tq6b/, put back from the head copy and checked put back, then lower and oracle whole.
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

target=internal/lower/library_object.go
head_copy="$TMPDIR/head_library_object.go"
# put_back restores the mutated file from its head copy and checks it's byte for byte the head again.
put_back() { cp "$head_copy" "$target"; if cmp -s "$head_copy" "$target"; then echo "PASS restored $target"; else echo "FAIL restore $target"; ok=0; fi; }
# mutate puts a whole-file mutant in place, refused unless it differs from head in exactly one line, so a mutant cut
# against an older head never carries that head's other lines back in.
mutate() {
	mutant=scratch-tq6b/$1
	changed=$(diff "$head_copy" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $1 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "$target"
	if cmp -s "$mutant" "$target"; then echo "PASS mutant $1 in place"; else echo "FAIL mutant $1 not in place"; ok=0; return 1; fi
}

cp "$target" "$head_copy"
# A killed run still leaves the tree as proved.
trap 'cp "$head_copy" "$target"' EXIT

# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
step gofmt bash -c 'difference=$(gofmt -d internal/lower/library_object.go internal/lower/representation_reason_test.go internal/oracle/intersection_optional_boolean_test.go); printf "%s\n" "$difference"; test -z "$difference"'
step vet go vet ./internal/lower/ ./internal/oracle/
step products sandboxed_products ./internal/oracle

# The pins at head, each program's verdict on its own line.
oracle_pins='TestIntersectionOptionalBoolean'
lower_pins='TestObjectIntersectionOptionalBoolean|TestRepresentationReasonStops'
step "oracle pins" go test -count=1 -run "$oracle_pins" ./internal/oracle/ -v
subtests
step "lower pins" go test -count=1 -run "$lower_pins" ./internal/lower/ -v
subtests

# The Record trace: each probe must stop at compile time on both backends, a record never lowered as the
# intersection's fixed object. Each probe's full stop is printed from its own log.
logs="$TMPDIR/probe-logs"
step "record probes" bash cloud/probe/run.sh scratch-tq6b/probes --logs "$logs"
cp "$TMPDIR/step.log" "$TMPDIR/probes.table"
for probe in scratch-tq6b/probes/*.ts; do
	name=$(basename "$probe")
	row=$(grep -F "| $name |" "$TMPDIR/probes.table")
	stops=$(printf '%s' "$row" | grep -o '| stops: ' | wc -l | tr -d ' ')
	node=$(tr '\n' ' ' < "$logs/$name/node.stdout")
	said=$(head -1 "$logs/$name/head.native-compile.stderr")
	if [ "$stops" = 2 ]; then echo "PASS probe $name stops on both backends (Node: $node): $said"; else echo "FAIL probe $name doesn't stop on both backends: $row"; ok=0; fi
done

# M1: member agreement dropped, so an intersection whose own field is an optional boolean is admitted whatever its
# members hold. disagreeing.ts lowers instead of stopping, and the lower pin's union and unknown members are admitted.
if mutate library_object.go.agreement.mutant; then
	red "M1 member agreement dropped, oracle" go test -count=1 -run 'TestIntersectionOptionalBooleanStops/^disagreeing\.ts$' ./internal/oracle/ -v -- '--- FAIL: TestIntersectionOptionalBooleanStops/disagreeing.ts'
	red "M1 member agreement dropped, lower" go test -count=1 -run 'TestObjectIntersectionOptionalBoolean|TestRepresentationReasonStops/^optional_boolean$' ./internal/lower/ -v -- '--- FAIL: TestObjectIntersectionOptionalBoolean/union_member' '--- FAIL: TestObjectIntersectionOptionalBoolean/unknown_member' '--- FAIL: TestRepresentationReasonStops/optional_boolean'
fi
put_back

# M2: the plain rule swapped back to the record rule (today's code), so every optional boolean field stops.
if mutate library_object.go.record_rule.mutant; then
	red "M2 record rule, oracle" go test -count=1 -run 'TestIntersectionOptionalBooleanAgreesWithNode' ./internal/oracle/ -v -- '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/read.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/declared_by_one.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/agreeing_members.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/inherited.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/optional_true.ts' "is an optional boolean, which its record storage doesn't hold"
	red "M2 record rule, lower" go test -count=1 -run 'TestObjectIntersectionOptionalBoolean' ./internal/lower/ -v -- '--- FAIL: TestObjectIntersectionOptionalBoolean/agreeing' '--- FAIL: TestObjectIntersectionOptionalBoolean/inherited'
fi
put_back

# M3: members asked by their own declarations instead of the checker's property, so a field a member has only
# through its base chain is found in no member, and inherited.ts stops.
if mutate library_object.go.declarations.mutant; then
	red "M3 own declarations, oracle" go test -count=1 -run 'TestIntersectionOptionalBooleanAgreesWithNode/^inherited\.ts$' ./internal/oracle/ -v -- '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/inherited.ts' "field isUnterminated is an optional boolean"
	red "M3 own declarations, lower" go test -count=1 -run 'TestObjectIntersectionOptionalBoolean/^inherited$' ./internal/lower/ -v -- '--- FAIL: TestObjectIntersectionOptionalBoolean/inherited'
fi
put_back
trap - EXIT

whole "lower whole" go test -count=1 ./internal/lower/
whole "oracle whole" go test -count=1 ./internal/oracle/

[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
