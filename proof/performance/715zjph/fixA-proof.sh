# #715zjph fix A: load.IsLibrary's prefix read once. vet, internal/load with its pin and the pin's mutant, then
# internal/lower and internal/oracle whole. Run after environment.sh, sandboxed-job.sh, sandboxed_npm stage3/api and
# sandboxed_products ./internal/lower ./internal/oracle. Never merged.
set -u
bad=0
step() { local name=$1; shift; local start=$(date +%s); "$@" > "$TMPDIR/a-$name.log" 2>&1; local code=$?; echo "$([ $code = 0 ] && echo PASS || echo FAIL) $name exit $code, $(( $(date +%s) - start )) s"; [ $code = 0 ] || { grep -E '^(--- FAIL|FAIL|panic|.*_test.go:[0-9]+:)' "$TMPDIR/a-$name.log" | head -10 | cut -c1-300; bad=1; }; return $code; }
echo "on $(hostname), $(nproc) cores, load $(cut -d' ' -f1-3 /proc/loadavg)"
step gofmt bash -c '[ -z "$(gofmt -l internal/load/)" ]'
step vet go vet ./internal/load/
step load go test -count=1 ./internal/load/
F=internal/load/load.go
cp "$F" "$TMPDIR/load.go"
python3 - "$F" <<'PYTHON'
import sys
p = sys.argv[1]; s = open(p).read(); old = 'sync.OnceValue(func() string { return bundled.LibPath().AsString() })'
assert s.count(old) == 1
open(p, 'w').write(s.replace(old, 'sync.OnceValue(func() string { return "" })'))
PYTHON
go test -count=1 -run '^TestLibraryPrefixIsTheLibraryPath$' ./internal/load/ > "$TMPDIR/a-m.log" 2>&1 && { echo "FAIL mutant: the pin stayed green"; bad=1; } || { grep -q -- '--- FAIL: TestLibraryPrefixIsTheLibraryPath' "$TMPDIR/a-m.log" && echo "PASS mutant: TestLibraryPrefixIsTheLibraryPath red" || { echo "FAIL mutant: red for another reason"; tail -3 "$TMPDIR/a-m.log"; bad=1; }; }
cp "$TMPDIR/load.go" "$F"
step lower go test -count=1 -timeout 30m ./internal/lower/
start=$(date +%s); go test -count=1 -timeout 30m ./internal/oracle/ > "$TMPDIR/o.log" 2>&1; echo "oracle exit $?, $(( $(date +%s) - start )) s"
other=$(grep -oE '^--- FAIL: [A-Za-z0-9_]+ ' "$TMPDIR/o.log" | cut -d' ' -f3 | sort -u | grep -vE '^(TestAssignmentProof|TestStringLengthJoinMemory$)' | tr '\n' ' ')
[ -z "$other" ] && echo "PASS oracle: only main's sealed set" || { echo "FAIL oracle: $other"; bad=1; }
echo "VERDICT $([ $bad = 0 ] && echo OK || echo FAIL)"
exit $bad
