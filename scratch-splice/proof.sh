set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
# #p1qw4kf, splice with a spread among its items. PART=a: gofmt, vet, TestSpliceSpread verbose, the four mutants of
# splice_spread/mutants.md each planted by python, run and restored, and the identity probe (every oracle program
# with a splice, push or unshift lowered at main's two files and at head, its C and JavaScript compared by sha256).
# PART=b: internal/lower whole. PART=c: internal/oracle whole, internal/native whole under WASI. PART=d: internal/ir,
# stage3/census/latent/..., internal/compilercensus, cmd/adamic-gate/parallelcensus, each whole.
part=${PART:-a}
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== " "$TMPDIR/step.log" | tail -40; ok=0; fi; }
whole() { name=$1; shift; begun=$SECONDS; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name in $((SECONDS - begun)) s"; else echo "FAIL $name (exit $code) in $((SECONDS - begun)) s"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
subtests() { sed -nE 's/^ *--- (PASS|FAIL|SKIP): ([^ ]+).*/\1 \2/p' "$TMPDIR/step.log"; grep -q "^ *--- SKIP" "$TMPDIR/step.log" && ok=0; }
head_files="internal/lower/array_call_arguments.go internal/lower/object.go"
keep() { for file in $head_files; do cp "$file" "$TMPDIR/$(basename "$file").head"; done; }
restore() {
	for file in $head_files; do
		cp "$TMPDIR/$(basename "$file").head" "$file"
		if cmp -s "$TMPDIR/$(basename "$file").head" "$file"; then echo "PASS restored $file after $1"; else echo "FAIL restore $file after $1"; ok=0; fi
	done
}
main_files() { cp scratch-splice/array_call_arguments.go.main internal/lower/array_call_arguments.go && cp scratch-splice/object.go.main internal/lower/object.go; }

echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
step npm sandboxed_npm stage3/api stage3/typescript

if [ "$part" = a ]; then
	step gofmt bash -c 'difference=$(gofmt -d internal/lower/array_call_arguments.go internal/lower/object.go internal/oracle/splice_spread_test.go); printf "%s\n" "$difference"; test -z "$difference"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	step products sandboxed_products ./internal/oracle
	step "splice spread" go test -count=1 -run '^TestSpliceSpread$' ./internal/oracle/ -v
	subtests
	grep -E "splice_spread_test.go:[0-9]+:" "$TMPDIR/step.log" | grep -v "gate cache" | head -20 | cut -c1-600
	keep
	for mutant in M1 M2 M3 M4; do
		if python3 -I scratch-splice/mutate.py "$mutant"; then
			go test -count=1 -run '^TestSpliceSpread$' ./internal/oracle/ -v > "$TMPDIR/mutant.log" 2>&1
			code=$?
			reds=$(sed -nE 's/^ *--- FAIL: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
			greens=$(sed -nE 's/^ *--- PASS: ([^ ]+).*/\1/p' "$TMPDIR/mutant.log" | tr '\n' ' ')
			if [ $code -eq 0 ]; then echo "FAIL mutant $mutant SURVIVED"; ok=0; else echo "PASS mutant $mutant RED: $reds"; fi
			echo "mutant $mutant green: $greens"
			grep -E "splice_spread_test.go:[0-9]+:" "$TMPDIR/mutant.log" | grep -v "gate cache" | head -12 | cut -c1-900
			grep -qE "^(--- FAIL|    --- FAIL|ok )" "$TMPDIR/mutant.log" || tail -20 "$TMPDIR/mutant.log"
		else
			echo "FAIL mutant $mutant: not planted"; ok=0
		fi
		restore "mutant $mutant"
	done
	# The identity probe: a splice without a spread, and push and unshift through the shared packing, lower to the
	# same C and JavaScript at main's two files as at head.
	grep -lE "\.(splice|unshift|push)\(" internal/oracle/testdata/*.a internal/oracle/testdata/*/*.a internal/oracle/testdata/*.ts internal/oracle/testdata/*/*.ts 2>/dev/null | grep -v splice_spread > "$TMPDIR/identity.list"
	echo "identity probe: $(wc -l < "$TMPDIR/identity.list") programs"
	cp scratch-splice/identity_probe_test.go.txt internal/oracle/zz_splice_identity_probe_test.go
	for side in head main; do
		[ "$side" = main ] && main_files
		SPLICE_IDENTITY_LIST="$TMPDIR/identity.list" go test -count=1 -timeout 60m -run '^TestZZSpliceIdentityProbe$' ./internal/oracle/ -v > "$TMPDIR/identity-$side.log" 2>&1
		echo "identity probe at $side: exit $?, $(grep -c '^PROBE' "$TMPDIR/identity-$side.log") lines, $(grep -c '^PROBE .* c ' "$TMPDIR/identity-$side.log") lowered, $(grep -c '^PROBE .* stop ' "$TMPDIR/identity-$side.log") stopped"
		grep '^PROBE' "$TMPDIR/identity-$side.log" > "$TMPDIR/identity-$side.lines"
		grep -q '^PROBE' "$TMPDIR/identity-$side.log" || { tail -30 "$TMPDIR/identity-$side.log"; ok=0; }
		restore "the identity probe at $side"
	done
	# The stop's words beside push's, and other representations, each lowered with its stop printed whole.
	ls scratch-splice/stops/*.ts internal/oracle/testdata/splice_spread/stop_representation.ts > "$TMPDIR/stops.list"
	SPLICE_PROBE_TEXT=1 SPLICE_IDENTITY_LIST="$TMPDIR/stops.list" go test -count=1 -run '^TestZZSpliceIdentityProbe$' ./internal/oracle/ -v > "$TMPDIR/stops.log" 2>&1
	echo "stops probe: exit $?"
	grep '^PROBE' "$TMPDIR/stops.log" | cut -c1-400
	rm internal/oracle/zz_splice_identity_probe_test.go
	if cmp -s "$TMPDIR/identity-head.lines" "$TMPDIR/identity-main.lines"; then
		echo "PASS identity: head and main byte-identical over $(wc -l < "$TMPDIR/identity-head.lines") programs, sha256 $(sha256sum "$TMPDIR/identity-head.lines" | cut -c1-64)"
	else
		echo "FAIL identity: head and main differ"; ok=0
		diff "$TMPDIR/identity-main.lines" "$TMPDIR/identity-head.lines" | head -20
	fi
	git status --porcelain
fi
if [ "$part" = b ]; then
	step "typescript source" sandboxed_typescript_source
	whole "internal/lower whole" go test -count=1 -timeout 80m ./internal/lower/
fi
if [ "$part" = c ]; then
	step products sandboxed_products ./internal/oracle
	whole "internal/oracle whole" go test -count=1 -timeout 80m ./internal/oracle/
	ADAMIC_TEST_WASI=1 whole "internal/native whole (ADAMIC_TEST_WASI=1)" go test -count=1 -timeout 80m ./internal/native/
fi
if [ "$part" = d ]; then
	step "typescript source" sandboxed_typescript_source
	whole "internal/ir whole" go test -count=1 -timeout 40m ./internal/ir/
	whole "stage3/census/latent/... whole" go test -count=1 -timeout 40m ./stage3/census/latent/...
	whole "internal/compilercensus whole" go test -count=1 -timeout 40m ./internal/compilercensus/
	whole "cmd/adamic-gate/parallelcensus whole" go test -count=1 -timeout 40m ./cmd/adamic-gate/parallelcensus/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
