set -u
# #rnbvya6's proof. PART=a: gofmt, vet, the pins at head, the mutant red, the C at base for return.ts.
# PART=b: internal/lower, internal/oracle, compilercensus and parallelcensus whole.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
base=358b95d57e3d6ed01b09ff3c32421c14bb3c2ee8
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; tail -40 "$TMPDIR/step.log"; ok=0; fi; }
want_red() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -ne 0 ]; then echo "PASS $name: red, as wanted"; grep -E "^\s+--- FAIL|error:|disagree|exit" "$TMPDIR/step.log" | head -12; else echo "FAIL $name: green, a survivor"; ok=0; fi; }
if [ "${PART}" = a ]; then
	step gofmt bash -c 'test -z "$(gofmt -l internal/lower/narrowed.go internal/oracle/narrowed_never_test.go)"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	sandboxed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
	step pins go test -count=1 -run 'TestNarrowedNeverReadsAgreeWithNode' ./internal/oracle/ -v
	grep -v "gate cache\|oracle_test.go:6[0-9][0-9]\|oracle_test.go:7[0-9][0-9]\|^=== " "$TMPDIR/step.log" | head -60
	cp internal/lower/narrowed.go "$TMPDIR/narrowed.go.head"
	cp scratch-rnbfix/narrowed.go.main internal/lower/narrowed.go; grep -q onlyAbsent internal/lower/narrowed.go && { echo "FAIL mutant not applied"; ok=0; }
	want_red "mutant (narrowed.go at main)" go test -count=1 -run 'TestNarrowedNeverReadsAgreeWithNode' ./internal/oracle/ -v
	cp "$TMPDIR/narrowed.go.head" internal/lower/narrowed.go
	if grep -q onlyAbsent internal/lower/narrowed.go; then echo "PASS restored"; else echo "FAIL restore"; ok=0; fi
fi
if [ "${PART}" = b ]; then
	sandboxed_products ./internal/lower ./internal/oracle ./internal/compilercensus > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
	step "lower whole" go test -count=1 ./internal/lower/
	step "oracle whole" go test -count=1 ./internal/oracle/
	step "compilercensus whole" go test -count=1 ./internal/compilercensus/
	step "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
