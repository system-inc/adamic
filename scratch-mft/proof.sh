set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #mft3zc6: sameRecordStorage judges an intersection of plain shapes as the fixed object objectIntersection holds
# it as, so a record seen as one stops at the relation in recordStorageView's words. PART=a: gofmt, vet, the pins
# at head with each subtest's verdict, the mutant restoring the nil answer from a whole-file copy in scratch-mft/,
# put back from the head copy and checked put back, then lower whole. PART=b: oracle whole and native whole
# (ADAMIC_TEST_WASI=1), so each part fits one job.
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

target=internal/lower/records.go
head_copy="$TMPDIR/head_records.go"
# put_back restores the mutated file from its head copy and checks it's byte for byte the head again.
put_back() { cp "$head_copy" "$target"; if cmp -s "$head_copy" "$target"; then echo "PASS restored $target"; else echo "FAIL restore $target"; ok=0; fi; }
# mutate puts a whole-file mutant in place, refused unless it differs from head in exactly one line, so a mutant cut
# against an older head never carries that head's other lines back in.
mutate() {
	mutant=scratch-mft/$1
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
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/records.go internal/lower/records_intersection_target_test.go internal/oracle/record_seen_as_intersection_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/lower ./internal/oracle

	# The pins at head, each program's verdict on its own line.
	step "oracle pins" go test -count=1 -run 'TestRecordSeenAsIntersection' ./internal/oracle/ -v
	subtests
	step "lower pin" go test -count=1 -run 'TestSameRecordStorageIntersectionTarget' ./internal/lower/ -v
	subtests

	# M1: the nil answer restored, so an intersection target is the same storage as anything. partial_number.ts
	# lowers again instead of stopping (natively it panics as a compiler bug, main's road), and the lower pin finds
	# a record, an index signature and a record field the same storage as the fixed object.
	if mutate records.go.nil_answer.mutant; then
		red "M1 nil answer, oracle" go test -count=1 -run 'TestRecordSeenAsIntersectionStops/^partial_number\.ts$' ./internal/oracle/ -v -- '--- FAIL: TestRecordSeenAsIntersectionStops/partial_number.ts' 'got <nil>'
		red "M1 nil answer, lower" go test -count=1 -run 'TestSameRecordStorageIntersectionTarget' ./internal/lower/ -v -- '--- FAIL: TestSameRecordStorageIntersectionTarget/partial_record' '--- FAIL: TestSameRecordStorageIntersectionTarget/index_signature' '--- FAIL: TestSameRecordStorageIntersectionTarget/nested_record_field'
	fi
	put_back
	trap - EXIT

	whole "lower whole" go test -count=1 ./internal/lower/
fi

if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle ./internal/native
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
fi

if [ "$part" != a ] && [ "$part" != b ]; then echo "FAIL PART=$part names no part, want a or b"; ok=0; fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
