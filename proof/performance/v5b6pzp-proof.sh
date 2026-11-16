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
sed -i 's/return shapeFitPromiseStop$/return ""/' "$F"
if cmp -s "$F" "$TMPDIR/v5-original"; then echo "FAIL mutant didn't apply"; bad=1; else
  go test -count=1 -timeout 40m -v -run '^TestPromiseIsNotYetWhereAnObjectIsHeld$' ./internal/oracle/ > "$TMPDIR/v5-mutant.log" 2>&1
  if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/v5-mutant.log"; then echo "FAIL mutant doesn't compile"; bad=1; else
    reds=$(grep -cE '^ *--- FAIL: TestPromiseIsNotYetWhereAnObjectIsHeld/(initializer|argument|union_slot|union_then_in|literal_field|literal_then_in|union_object_keyword)\.(a|ts) ' "$TMPDIR/v5-mutant.log")
    others=$(grep -cE '^ *--- FAIL: TestPromiseIsNotYetWhereAnObjectIsHeld/(typed_promise|union_promise_member|awaited|unknown_slot|promise_all|resolve_adopts)' "$TMPDIR/v5-mutant.log")
    echo "$([ "$reds" = 14 ] && [ "$others" = 0 ] && echo PASS || echo FAIL) mutant: $reds of 14 probes red, $others others red"; { [ "$reds" = 14 ] && [ "$others" = 0 ]; } || bad=1
    grep -E '_test.go:[0-9]+:' "$TMPDIR/v5-mutant.log" | head -2 | cut -c1-300
  fi
fi
cp "$TMPDIR/v5-original" "$F"
# Mutant: the async return exemption dropped, so the stop fires first at an async return of a promise.
sed -i 's/if l.function != nil \&\& l.function.Async \&\& parent != nil/if false \&\& parent != nil/' "$F"
if cmp -s "$F" "$TMPDIR/v5-original"; then echo "FAIL return mutant didn't apply"; bad=1; else
  go test -count=1 -timeout 40m -v -run '^TestPromiseIsNotYetWhereAnObjectIsHeld$' ./internal/oracle/ > "$TMPDIR/v5-return.log" 2>&1
  if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/v5-return.log"; then echo "FAIL return mutant doesn't compile"; bad=1; else
    reds=$(grep -cE '^ *--- FAIL: TestPromiseIsNotYetWhereAnObjectIsHeld/async_(return|arrow)_adopts\.(a|ts) ' "$TMPDIR/v5-return.log")
    echo "$([ "$reds" -ge 2 ] && echo PASS || echo FAIL) return mutant: $reds of 4 adoption pins red"; [ "$reds" -ge 2 ] || bad=1
    grep -E 'adopts.*_test.go:[0-9]+:|want it stopped' "$TMPDIR/v5-return.log" | head -2 | cut -c1-300
  fi
fi
cp "$TMPDIR/v5-original" "$F"
# Mutant: NonPrimitive dropped from the masks, so a number | object union boxes the promise.
cp "$F" "$TMPDIR/v5-original"
sed -i 's/|checker.TypeFlagsNonPrimitive) != 0/) != 0/' "$F"
if cmp -s "$F" "$TMPDIR/v5-original"; then echo "FAIL keyword mutant didn't apply"; bad=1; else
  go test -count=1 -timeout 40m -v -run '^TestPromiseIsNotYetWhereAnObjectIsHeld$' ./internal/oracle/ > "$TMPDIR/v5-keyword.log" 2>&1
  if grep -qE 'build failed|declared and not used|\[setup failed\]' "$TMPDIR/v5-keyword.log"; then echo "FAIL keyword mutant doesn't compile"; bad=1; else
    reds=$(grep -cE '^ *--- FAIL: TestPromiseIsNotYetWhereAnObjectIsHeld/union_object_keyword\.(a|ts) ' "$TMPDIR/v5-keyword.log")
    echo "$([ "$reds" = 2 ] && echo PASS || echo FAIL) keyword mutant: $reds of 2 red"; [ "$reds" = 2 ] || bad=1
    grep -E 'union_object_keyword.*_test.go:[0-9]+:|want NotYet' "$TMPDIR/v5-keyword.log" | head -2 | cut -c1-360
  fi
fi
cp "$TMPDIR/v5-original" "$F"
step oracle go test -count=1 -timeout 40m ./internal/oracle/
echo "oracle top-level fails: $(grep -oE "^--- FAIL: [A-Za-z0-9_]+ " "$TMPDIR/v5-oracle.log" | cut -d" " -f3 | sort -u | tr "\n" " ")"
step lower go test -count=1 -timeout 40m ./internal/lower/
step fixtures go test -count=1 -timeout 40m ./stage3/fixtures/...
step parallelcensus go test -count=1 ./cmd/adamic-gate/parallelcensus/
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
