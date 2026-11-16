# #v5b6pzp's proof, sealed: the parent of this commit is the head. Run after environment.sh, sealed-job.sh and
# sealed_products ./internal/oracle ./internal/lower. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/v5-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/v5-$name.log" | head -14 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/lower/ internal/oracle/)" ]'
step vet go vet ./internal/lower/ ./internal/oracle/
step pins go test -count=1 -timeout 40m -v -run '^TestPromiseIsNotYetWhereAnObjectIsHeld$' ./internal/oracle/
grep -E '^ *--- (PASS|FAIL)|stops as on main' "$TMPDIR/v5-pins.log" | sed 's/^ *//' | cut -c1-260 | head -30
# Mutant: the stop answers "", so both probes lower again.
F=internal/lower/shape_fit.go
cp "$F" "$TMPDIR/v5-original"
sed -i 's/^\treturn shapeFitPromiseStop$/\treturn ""/' "$F"
if cmp -s "$F" "$TMPDIR/v5-original"; then echo "FAIL mutant didn't apply"; bad=1; else
  go test -count=1 -timeout 40m -v -run '^TestPromiseIsNotYetWhereAnObjectIsHeld$' ./internal/oracle/ > "$TMPDIR/v5-mutant.log" 2>&1
  if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/v5-mutant.log"; then echo "FAIL mutant doesn't compile"; bad=1; else
    reds=$(grep -cE '^ *--- FAIL: TestPromiseIsNotYetWhereAnObjectIsHeld/(initializer|argument)\.(a|ts) ' "$TMPDIR/v5-mutant.log")
    others=$(grep -cE '^ *--- FAIL: TestPromiseIsNotYetWhereAnObjectIsHeld/(typed_promise|union_slot|literal_field|unknown_slot|promise_all|resolve_adopts)' "$TMPDIR/v5-mutant.log")
    echo "$([ "$reds" = 4 ] && [ "$others" = 0 ] && echo PASS || echo FAIL) mutant: $reds of 4 probes red, $others others red"; { [ "$reds" = 4 ] && [ "$others" = 0 ]; } || bad=1
    grep -E '_test.go:[0-9]+:' "$TMPDIR/v5-mutant.log" | head -2 | cut -c1-300
  fi
fi
cp "$TMPDIR/v5-original" "$F"
step oracle go test -count=1 -timeout 40m ./internal/oracle/
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 40m ./stage3/fixtures/...
step parallelcensus go test -count=1 ./cmd/adamic-gate/parallelcensus/
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
