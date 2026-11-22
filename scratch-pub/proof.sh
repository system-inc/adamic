set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #sx8pbw1's publish gap on main f7fe5ce5: a field a spread's literal writes into a slot its staged source reserved
# absent is published and made ready, in objectLiteral's spread arm (internal/native/emit_objects.go) and on the
# reused path (internal/native/reuse.go), so it reads, enumerates and stringifies as written.
# PART=a: gofmt, vet, products, the pins at head with each subtest's verdict, each mutant from a whole-file copy in
# scratch-pub/, put back from the head copy and checked put back, then native whole (ADAMIC_TEST_WASI=1) and
# compilercensus whole.
# PART=b: oracle whole, so each part fits one job.
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

# put_back restores a mutated file from its head copy and checks it's byte for byte the head again.
put_back() { cp "$TMPDIR/head_$(basename "$1")" "$1"; if cmp -s "$TMPDIR/head_$(basename "$1")" "$1"; then echo "PASS restored $1"; else echo "FAIL restore $1"; ok=0; fi; }
# mutate puts a whole-file mutant of a target in place, refused unless it differs from head in exactly one line, so
# a mutant cut against an older head never carries that head's other lines back in.
mutate() {
	target=$1 mutant=scratch-pub/$2
	changed=$(diff "$TMPDIR/head_$(basename "$target")" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $2 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "$target"
	if cmp -s "$mutant" "$target"; then echo "PASS mutant $2 in place"; else echo "FAIL mutant $2 not in place"; ok=0; return 1; fi
}

echo "box: $(hostname) at $(git rev-parse HEAD) PART=$part"
# The whole runs read stage3's TypeScript (TestAssignmentProof*), which the box lays only on asking.
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = a ]; then
	cp internal/native/emit_objects.go "$TMPDIR/head_emit_objects.go"
	cp internal/native/reuse.go "$TMPDIR/head_reuse.go"
	# A killed run still leaves the tree as proved.
	trap 'cp "$TMPDIR/head_emit_objects.go" internal/native/emit_objects.go; cp "$TMPDIR/head_reuse.go" internal/native/reuse.go' EXIT

	# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
	step gofmt bash -c 'difference=$(gofmt -d internal/native/emit_objects.go internal/native/reuse.go internal/oracle/spread_staged_override_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/native/ ./internal/oracle/
	step products sandboxed_products ./internal/oracle ./internal/native

	# The pins at head, each program's verdict on its own line.
	step "oracle pins" go test -count=1 -run 'TestSpreadStagedOverride' ./internal/oracle/ -v
	subtests

	# M1: the publish removed from objectLiteral's spread arm, so the copy path's programs read the field absent
	# again: g01, g03 and g08 print undefined and drop it from keys and JSON, g04 stops at its read.
	if mutate internal/native/emit_objects.go emit_objects.go.publish_removed.mutant; then
		red "M1 spread arm's publish removed" go test -count=1 -run 'TestSpreadStagedOverride' ./internal/oracle/ -v -- '--- FAIL: TestSpreadStagedOverride/g01_cast_spread_override.ts' '--- FAIL: TestSpreadStagedOverride/g03_factory_spread_override.ts' '--- FAIL: TestSpreadStagedOverride/g04_merged_class_spread_override.ts' '--- FAIL: TestSpreadStagedOverride/g08_nested_spread.ts' "read before assignment: field 'label' in m.label"
	fi
	put_back internal/native/emit_objects.go

	# M2: the publish removed from the reused path, so g06, whose spread takes over a local's object, reads pos
	# absent again.
	if mutate internal/native/reuse.go reuse.go.publish_removed.mutant; then
		red "M2 reused path's publish removed" go test -count=1 -run 'TestSpreadStagedOverride' ./internal/oracle/ -v -- '--- FAIL: TestSpreadStagedOverride/g06_reused_in_function.ts'
	fi
	put_back internal/native/reuse.go
	trap - EXIT

	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 -timeout 40m ./internal/native/
	step "products compilercensus" sandboxed_products ./internal/compilercensus
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
fi

if [ "$part" = b ]; then
	step products sandboxed_products ./internal/oracle
	whole "oracle whole" go test -count=1 ./internal/oracle/
fi

if [ "$part" != a ] && [ "$part" != b ]; then echo "FAIL PART=$part names no part, want a or b"; ok=0; fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
