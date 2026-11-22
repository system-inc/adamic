set -u
# #24kvb6k's Now: a contract with an optional field admits no class constructor. PART=a: gofmt, vet, the pins, the
# lower pin and the mutant. PART=b: lower, oracle, native (ADAMIC_TEST_WASI=1), compilercensus, parallelcensus whole.
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
ok=1
step() { name=$1; shift; "$@" > "$TMPDIR/step.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -v "gate cache\|^=== \|oracle_test.go:[67][0-9][0-9]" "$TMPDIR/step.log" | tail -60; ok=0; fi; }
whole() { name=$1; shift; "$@" > "$TMPDIR/whole.log" 2>&1; code=$?; if [ $code -eq 0 ]; then echo "PASS $name"; else echo "FAIL $name (exit $code)"; grep -E "^(--- FAIL|    --- FAIL|panic:)" "$TMPDIR/whole.log" | sort -u | head -40; ok=0; fi; }
if [ "$PART" = a ]; then
	step gofmt bash -c 'out=$(gofmt -d internal/lower/view_contracts.go internal/lower/view_statics_admitted_test.go internal/oracle/view_static_class_test.go); echo "$out"; test -z "$out"'
	step vet go vet ./internal/lower/ ./internal/oracle/
	sandboxed_products ./internal/oracle > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
	step "view_static_class pins" go test -count=1 -run 'TestViewStaticClass' ./internal/oracle/ -v
	step "lower pin" go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ -v
	cp internal/lower/view_contracts.go "$TMPDIR/view_contracts.go.head"
	cp scratch-now/view_contracts.go.mutant internal/lower/view_contracts.go
	if grep -q "field.Optional || l.objectPrototypeName" internal/lower/view_contracts.go; then echo "FAIL mutant not applied"; ok=0; fi
	go test -count=1 -run 'TestViewStaticClassStops' ./internal/oracle/ -v > "$TMPDIR/mutant.log" 2>&1
	if [ $? -ne 0 ]; then echo "PASS mutant (the optional condition dropped): pins red, as wanted"; grep -E "^\s+--- FAIL|differs" "$TMPDIR/mutant.log" | head -14; else echo "FAIL mutant survived the pins"; ok=0; fi
	go test -count=1 -run 'TestViewStaticsAdmittedAsksFunctionAndObject' ./internal/lower/ > "$TMPDIR/mutant2.log" 2>&1
	if [ $? -ne 0 ]; then echo "PASS mutant: lower pin red, as wanted"; else echo "FAIL mutant survived the lower pin"; ok=0; fi
	cp "$TMPDIR/view_contracts.go.head" internal/lower/view_contracts.go
	if grep -q "field.Optional || l.objectPrototypeName" internal/lower/view_contracts.go; then echo "PASS restored"; else echo "FAIL restore"; ok=0; fi
fi
if [ "$PART" = b ]; then
	sandboxed_npm stage3/api stage3/typescript > "$TMPDIR/npm.log" 2>&1 || { echo "GAP npm"; tail -8 "$TMPDIR/npm.log"; }
	sandboxed_products ./internal/lower ./internal/oracle ./internal/native ./internal/compilercensus > "$TMPDIR/products.log" 2>&1 || { echo "GAP products"; tail -20 "$TMPDIR/products.log"; }
	whole "lower whole" go test -count=1 ./internal/lower/
	whole "oracle whole" go test -count=1 ./internal/oracle/
	whole "native whole, ADAMIC_TEST_WASI=1" env ADAMIC_TEST_WASI=1 go test -count=1 ./internal/native/
	whole "compilercensus whole" go test -count=1 ./internal/compilercensus/
	whole "parallelcensus whole" go test -count=1 ./cmd/adamic-gate/parallelcensus/
fi
[ $ok -eq 1 ] && echo PASS || { echo FAIL; exit 1; }
