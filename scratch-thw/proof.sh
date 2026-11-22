set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #thwgcax: a logical whose right fit can't hold as its maybe result (a number | undefined field narrowed to
# undefined, under a boolean | undefined &&) stops at lowering with both representations named, where it once
# emitted C clang refused. gofmt, vet, the pins at head, then the mutant: the check removed from a whole-file copy in
# scratch-thw/, under which narrowed_number.ts lowers and goes red at clang, put back from the head copy and checked.
# Then lower and oracle whole.
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; ok=0; fi; }
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

expression=internal/lower/expression.go
head_copy="$TMPDIR/head_expression.go"
# put_back restores the mutated file from its head copy and checks it's byte for byte the head again.
put_back() { cp "$head_copy" "$expression"; if cmp -s "$head_copy" "$expression"; then echo "PASS restored $expression"; else echo "FAIL restore $expression"; ok=0; fi; }
# mutate puts the whole-file mutant in place, refused unless it differs from head only by the check's nine lines
# removed (the if, its six comment lines, the notYet return and the closing brace), so a mutant cut against an older
# head never carries that head's other lines back in.
mutate() {
	mutant=scratch-thw/expression.go.mutant
	removed=$(diff "$head_copy" "$mutant" | grep -c '^<')
	added=$(diff "$head_copy" "$mutant" | grep -c '^>')
	if [ "$removed" != 9 ] || [ "$added" != 0 ]; then echo "FAIL mutant differs from head by $removed removed and $added added lines, want 9 removed and 0 added"; ok=0; return 1; fi
	if ! diff "$head_copy" "$mutant" | grep -qF '< 			if fitted.Type() != of {'; then echo "FAIL mutant's removed lines aren't the check"; ok=0; return 1; fi
	cp "$mutant" "$expression"
	if cmp -s "$mutant" "$expression"; then echo "PASS mutant in place"; else echo "FAIL mutant not in place"; ok=0; return 1; fi
}

cp "$expression" "$head_copy"
# A killed run still leaves the tree as proved.
trap 'cp "$head_copy" "$expression"' EXIT

# gofmt -d prints any difference, so a FAIL shows what gofmt wants.
step gofmt bash -c 'difference=$(gofmt -d internal/lower/expression.go internal/oracle/logical_maybe_mismatch_test.go); printf "%s\n" "$difference"; test -z "$difference"'
step vet go vet ./internal/lower/ ./internal/oracle/
step products sandboxed_products ./internal/oracle

# The pins at head, each test's and program's verdict on its own line.
step "logical_maybe_mismatch pins" go test -count=1 -run '^TestLogicalMaybeMismatch' ./internal/oracle/ -v
subtests

# The check removed: narrowed_number.ts lowers again, and native's build of it is clang's refusal.
if mutate; then
	red "mutant: check removed, narrowed_number.ts" go test -count=1 -run '^TestLogicalMaybeMismatchStops$' ./internal/oracle/ -v -- 'clang failed' '--- FAIL: TestLogicalMaybeMismatchStops'
fi
put_back
trap - EXIT

sandboxed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 || echo "GAP npm"
whole "lower whole" go test -count=1 ./internal/lower/
whole "oracle whole" go test -count=1 ./internal/oracle/

[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
