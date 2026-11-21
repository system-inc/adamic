set -u
# #s03ndgw piece 1's broad gate: every package the head reaches, whole. Each line PASS or FAIL, every red named.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
sandboxed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 || { echo "GAP npm"; tail -8 "$TMPDIR/npm.log"; }
sandboxed_products ./internal/lower ./internal/oracle ./internal/native ./internal/javascript ./internal/compilercensus ./internal/ir > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
ok=1
whole() {
	name=$1; shift
	"$@" > "$TMPDIR/whole.log" 2>&1
	code=$?
	if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; ok=0; fi
}
whole "lower whole" go test -count=1 ./internal/lower/
whole "oracle whole" go test -count=1 ./internal/oracle/
whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 ./internal/native/
whole "javascript whole" go test -count=1 ./internal/javascript/
whole "ir whole (the call-target census)" go test -count=1 ./internal/ir/
whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
whole "statecopy whole" go test -count=1 ./stage3/census/latent/statecopy/
whole "census/latent whole" go test -count=1 ./stage3/census/latent/
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
