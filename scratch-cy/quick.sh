set -u
source /opt/adamic-block/environment.sh
source cloud/sandboxed-job.sh
export GOFLAGS="${GOFLAGS:+$GOFLAGS }-timeout=40m"
# #sbvnh96 quick read: the pins, the cycle equality family and the census equality, before the whole gates.
echo "box: $(hostname) at $(git rev-parse HEAD)"
sandboxed_npm stage3/api stage3/typescript > /dev/null 2>&1
sandboxed_typescript_source > /dev/null 2>&1
sandboxed_products ./internal/oracle ./internal/lower > /dev/null 2>&1
ok=1
go test -count=1 -v -run '^(TestFreshLiteralCyclesAreCollected|TestFreshLiteralToJSONReachStops|TestConcurrencyRefusals)$' ./internal/oracle/ > "$TMPDIR/pins.log" 2>&1 || ok=0
sed -nE 's/^ *--- (PASS|FAIL): ([^ ]+).*/\1 \2/p' "$TMPDIR/pins.log" | grep -v "TestConcurrencyRefusals/" ; grep -E "TestConcurrencyRefusals/hidden_(literal_field|subclass)" "$TMPDIR/pins.log" | head -4
grep -E "_test.go:[0-9]+:" "$TMPDIR/pins.log" | head -10
go test -count=1 -run '^(TestCycleFinderEquality|TestCycleFinderCensusEquality|TestCycleIndexCandidateEdgeEquality)' ./internal/lower/ > "$TMPDIR/eq.log" 2>&1 || ok=0
grep -E "^(--- FAIL|    --- FAIL|ok|FAIL)|difference:" "$TMPDIR/eq.log" | head -20
[ $ok -eq 1 ] && echo QUICK PASS || { echo QUICK FAIL; exit 1; }
