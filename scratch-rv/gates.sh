set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #hh1cezv: Seats' build22-162 round 4 and build22-55 round 3, applied together on main 0c9b1201 (d7b51390 is 162,
# 69ed1373 is 55 on it). PART=p: Emit's probes for each family through cloud/probe/run.sh, against main.
# PART=a: lower, ir, javascript, statecopy, compilercensus, parallelcensus and stage3/fixtures whole.
# PART=b: oracle whole and native whole (ADAMIC_TEST_WASI=1).
part=${PART:-a}
ok=1
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)|_test.go:[0-9]+:" "$TMPDIR/whole.log" | sort -u | head -60; grep -qE "^(--- FAIL|panic:)" "$TMPDIR/whole.log" || tail -25 "$TMPDIR/whole.log"; ok=0; fi; }
echo "box: $(hostname) at $(git rev-parse HEAD 2>/dev/null) PART=$part"
whole npm sandboxed_npm stage3/api stage3/typescript
if [ "$part" = p ]; then
	for family in p162 p55; do
		echo "== $family"
		bash cloud/probe/run.sh "scratch-rv/$family" --logs "$TMPDIR/logs-$family" > "$TMPDIR/table-$family.txt" 2>&1 || { echo "FAIL run.sh $family"; ok=0; }
		tail -40 "$TMPDIR/table-$family.txt"
		for log in "$TMPDIR/logs-$family"/*; do
			[ -f "$log" ] || continue
			echo "-- $(basename "$log")"
			head -c 1500 "$log"
			echo
		done
	done
fi
if [ "$part" = a ]; then
	whole "lower whole" go test -count=1 ./internal/lower/
	whole "ir whole" go test -count=1 ./internal/ir/
	whole "javascript whole" go test -count=1 ./internal/javascript/
	whole "statecopy whole" go test -count=1 ./stage3/census/latent/statecopy/
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
	whole "stage3/fixtures whole" go test -count=1 ./stage3/fixtures/
	whole "admission-delta's own tests" go test -count=1 ./cmd/adamic-admission-delta/
fi
if [ "$part" = b ]; then
	whole products sandboxed_products ./internal/oracle
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 ./internal/native/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
