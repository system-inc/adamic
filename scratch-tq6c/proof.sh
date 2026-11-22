set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #tq6yzqy on main 178b45d3 (after #mft3zc6): an intersection field held as an optional boolean by the plain rule,
# as every member that has it holds it, is admitted (objectIntersection, intersectionMembersHoldMaybeBoolean), and a
# finite Partial<Record<"valid", boolean>> seen as such an intersection stops at the relation (partial_boolean.ts).
# PART=a: gofmt, vet, the pins at head with each subtest's verdict, the Record trace's probes each stopping on both
# backends, each mutant from a whole-file copy in scratch-tq6c/, put back from the head copy and checked put back, then
# lower whole and compilercensus whole. PART=b: oracle whole and native whole (ADAMIC_TEST_WASI=1), so each part fits
# one job.
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
	mutant=scratch-tq6c/$1
	changed=$(diff "$head_copy" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $1 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "$target"
	if cmp -s "$mutant" "$target"; then echo "PASS mutant $1 in place"; else echo "FAIL mutant $1 not in place"; ok=0; return 1; fi
}

echo "box: $(hostname) at $(git rev-parse HEAD) PART=$part"
# The whole runs read stage3's TypeScript (TestAssignmentProof*), which the box lays only on asking.
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = a ]; then
	cp "$target" "$head_copy"
	# A killed run still leaves the tree as proved.
	trap 'cp "$head_copy" "$target"' EXIT

	# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/library_object.go internal/lower/representation_reason_test.go internal/lower/object_intersection_fields_test.go internal/lower/map_union_values_test.go internal/oracle/intersection_optional_boolean_test.go internal/oracle/record_seen_as_intersection_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/oracle ./internal/compilercensus

	# The pins at head, each program's verdict on its own line.
	step "oracle pins" go test -count=1 -run 'TestIntersectionOptionalBoolean|TestRecordSeenAsIntersection' ./internal/oracle/ -v
	subtests
	step "lower pins" go test -count=1 -run 'TestObjectIntersectionOptionalBoolean|TestRepresentationReasonStops|TestObjectIntersectionFields|TestMapUnionValueBoundaries' ./internal/lower/ -v
	subtests

	# The Record trace: every road a record takes to the intersection stops on both backends, never read as the
	# intersection's fixed object. Each backend's stop is printed in its full first line from its log.
	logs="$TMPDIR/probe-logs"
	step "record probes" bash cloud/probe/run.sh scratch-tq6c/probes --logs "$logs"
	cp "$TMPDIR/step.log" "$TMPDIR/probes.table"
	for probe in scratch-tq6c/probes/*.ts; do
		name=$(basename "$probe")
		row=$(grep -F "| $name |" "$TMPDIR/probes.table")
		node=$(tr '\n' ' ' < "$logs/$name/node.stdout")
		if [ -z "$row" ]; then echo "FAIL probe $name has no row"; ok=0; continue; fi
		if printf '%s' "$row" | grep -qE '\| (agrees|differs: )'; then echo "FAIL probe $name: a backend doesn't stop (Node: $node): $row"; ok=0; continue; fi
		for backend in js native; do
			said=$(head -1 "$logs/$name/head.$backend-compile.stderr" 2>/dev/null)
			if [ -n "$said" ]; then echo "PASS probe $name $backend stops (Node: $node): $said"; else echo "FAIL probe $name $backend said nothing"; ok=0; fi
		done
	done

	# M1: member agreement dropped, so an intersection whose own field is an optional boolean is admitted whatever its
	# members hold. disagreeing.ts lowers instead of stopping, and the lower pins' union and unknown members are admitted.
	if mutate library_object.go.agreement.mutant; then
		red "M1 member agreement dropped, oracle" go test -count=1 -run 'TestIntersectionOptionalBooleanStops/^disagreeing\.ts$' ./internal/oracle/ -v -- '--- FAIL: TestIntersectionOptionalBooleanStops/disagreeing.ts'
		red "M1 member agreement dropped, lower" go test -count=1 -run 'TestObjectIntersectionOptionalBoolean|TestRepresentationReasonStops/^optional_boolean$' ./internal/lower/ -v -- '--- FAIL: TestObjectIntersectionOptionalBoolean/union_member' '--- FAIL: TestObjectIntersectionOptionalBoolean/unknown_member' '--- FAIL: TestRepresentationReasonStops/optional_boolean'
	fi
	put_back

	# M2: the plain rule swapped back to the record rule (main's code), so every optional boolean field stops. The
	# admitted programs stop, and partial_boolean.ts's intersection is no fixed object, so it stops in the
	# intersection's words instead of at the relation.
	if mutate library_object.go.record_rule.mutant; then
		red "M2 record rule, oracle" go test -count=1 -run 'TestIntersectionOptionalBooleanAgreesWithNode|TestRecordSeenAsIntersectionStops/^partial_boolean\.ts$' ./internal/oracle/ -v -- '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/read.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/declared_by_one.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/agreeing_members.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/inherited.ts' '--- FAIL: TestIntersectionOptionalBooleanAgreesWithNode/optional_true.ts' '--- FAIL: TestRecordSeenAsIntersectionStops/partial_boolean.ts' "is an optional boolean, which its record storage doesn't hold"
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
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
fi

if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle ./internal/native
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi

if [ "$part" != a ] && [ "$part" != b ]; then echo "FAIL PART=$part names no part, want a or b"; ok=0; fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
