set -u
# #24kvb6k's Now at 92554beb, the packages whole, split in two so each fits a Circle's 45m ceiling.
# PART=1: lower, compilercensus, parallelcensus. PART=2: oracle, native (ADAMIC_TEST_WASI=1).
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
ok=1
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; ok=0; fi; }
sandboxed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 || { echo "GAP npm"; tail -8 "$TMPDIR/npm.log"; }
if [ "$PART" = 1 ]; then
	sandboxed_products ./internal/lower ./internal/compilercensus > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
	whole "lower whole" go test -count=1 ./internal/lower/
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi
if [ "$PART" = 2 ]; then
	sandboxed_products ./internal/oracle ./internal/native > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 ./internal/native/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
