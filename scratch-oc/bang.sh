set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #cdpxjmb's bang_inside.ts (Language's o01, o04, o08) at 626b2fa4: gofmt, vet, the chain's oracle pins at head
# with each subtest's verdict, then M1 (steps unguarded) and M4 (steps before the guard) each red on
# bang_inside.ts, the file put back and checked between.
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -60; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log" | sed 's/^SKIP \(.*\)/FAIL \1 (skipped)/'; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
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
	grep -E "bang_inside" "$TMPDIR/mutant.log" | head -8
}
chain=internal/lower/optional_chain.go
chain_head="$TMPDIR/head_optional_chain.go"
cp "$chain" "$chain_head"
trap 'cp "$chain_head" "$chain"' EXIT
put_back() { cp "$chain_head" "$chain"; if cmp -s "$chain_head" "$chain"; then echo "PASS restored $chain"; else echo "FAIL restore $chain"; ok=0; fi; }
mutate() {
	mutant=scratch-oc/$1
	changed=$(diff "$chain_head" "$mutant" | grep -c '^[<>]')
	if [ "$changed" != 2 ]; then echo "FAIL mutant $1 differs from head in $changed lines, want one changed line"; ok=0; return 1; fi
	cp "$mutant" "$chain"
	if cmp -s "$mutant" "$chain"; then echo "PASS mutant $1 in place"; else echo "FAIL mutant $1 not in place"; ok=0; return 1; fi
}
pin=(go test -count=1 -run '^TestOptionalChainsAgreeWithNode$/^bang_inside.ts$' ./internal/oracle/ -v)

echo "box: $(hostname) at $(git rev-parse HEAD)"
step gofmt bash -c 'difference=$(gofmt -d internal/oracle/optional_chain_test.go); printf "%s\n" "$difference"; test -z "$difference"'
step vet go vet ./internal/oracle/
step products sandboxed_products ./internal/oracle
step "oracle pins" go test -count=1 -run '^(TestOptionalChainsAgreeWithNode|TestOptionalChainAssertionsKeepTheirChecks)$' ./internal/oracle/ -v
subtests
if mutate optional_chain.go.m1_steps_unguarded.mutant; then
	red "M1 steps unguarded" "${pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/bang_inside.ts'
fi
put_back
if mutate optional_chain.go.m4_steps_before_guard.mutant; then
	red "M4 steps before the guard" "${pin[@]}" -- '--- FAIL: TestOptionalChainsAgreeWithNode/bang_inside.ts'
fi
put_back
if [ $ok = 1 ]; then echo "PROOF PASS"; else echo "PROOF FAIL"; exit 1; fi
